import 'dart:collection';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:universal_html/html.dart' as html;
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';
import 'package:csv/csv.dart';
import 'package:path_provider/path_provider.dart';

const String extraColumn = "Add new column";
const String sumStr = "Somme()";
const String avgStr = "Moyenne()";

class ResultsPage extends StatefulWidget {
  List<String> selectedAttrs;
  List<String> selectedObjects;
  final String dateRange;
  final String namespace;

  ResultsPage(
      {super.key,
      required this.selectedAttrs,
      required this.selectedObjects,
      required this.dateRange,
      required this.namespace});

  @override
  State<StatefulWidget> createState() => _ResultsPageState();
}

class _ResultsPageState extends State<ResultsPage> {
  final SplayTreeSet<String> _allAttributes = SplayTreeSet<String>(
      (a, b) => a.toLowerCase().compareTo(b.toLowerCase()));
  Map<String, Map<String, dynamic>>? _data;
  List<DataColumn> columnLabels = [];
  List<String> selectedAttrs = [];

  // TODO: IMPLEMENT SORT
  bool sort = true;
  int sortColumnIndex = -1;
  // List<Data>? filterData;
  onsortColum(int columnIndex, bool ascending) {
    if (columnIndex == 0) {
      if (ascending) {
        // filterData!.sort((a, b) => a.name!.compareTo(b.name!));
      } else {
        // filterData!.sort((a, b) => b.name!.compareTo(a.name!));
      }
    }
  }

  @override
  void initState() {
    selectedAttrs = widget.selectedAttrs;
    if (!selectedAttrs.contains(extraColumn)) {
      selectedAttrs.add(extraColumn);
    }
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    bool isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return FutureBuilder(
        future: _data == null ? getData() : null,
        builder: (context, _) {
          if (_data != null) {
            if (_data!.isEmpty) {
              return SizedBox(
                height: MediaQuery.of(context).size.height > 205
                    ? MediaQuery.of(context).size.height - 205
                    : MediaQuery.of(context).size.height,
                child: Card(
                  margin: const EdgeInsets.all(0.1),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.warning_rounded,
                        size: 50,
                        color: Colors.grey.shade600,
                      ),
                      Padding(
                        padding: const EdgeInsets.only(top: 16),
                        child: Text(
                            "${AppLocalizations.of(context)!.noObjectsFound} :("),
                      ),
                    ],
                  ),
                ),
              );
            }
            return SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: PaginatedDataTable(
                header: Text(
                  localeMsg.yourReport,
                  style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: isSmallDisplay ? 16 : null),
                ),
                actions: [
                  Padding(
                    padding: const EdgeInsets.only(right: 6.0),
                    child: IconButton(
                      icon: const Icon(Icons.file_download_outlined),
                      onPressed: () => getCSV(),
                    ),
                  ),
                  PopupMenuButton<String>(
                    tooltip: localeMsg.selectionOptions,
                    offset: const Offset(0, -32),
                    itemBuilder: (_) => attributesCheckList(selectedAttrs),
                    onCanceled: () => print('canceled'),
                    icon: const Icon(Icons.add),
                  ),
                  PopupMenuButton<String>(
                    tooltip: localeMsg.mathFuncTip,
                    offset: const Offset(0, -32),
                    itemBuilder: (_) => mathFunctionsPopup(),
                    onCanceled: () => print('canceled'),
                    icon: const Icon(Icons.calculate_outlined),
                  )
                ],
                rowsPerPage: widget.selectedObjects.length >= 15
                    ? 15
                    : widget.selectedObjects.length,
                sortColumnIndex: sortColumnIndex > 0 ? sortColumnIndex : null,
                sortAscending: sort,
                columns: columnLabels,
                showCheckboxColumn: false,
                source: _DataSource(
                    context, selectedAttrs, widget.selectedObjects, _data),
              ),
            );
          } else {
            return const Center(child: CircularProgressIndicator());
          }
        });
  }

  getData() async {
    print("GET DATA");
    if (widget.namespace == Namespace.Test.name) {
      _data = getSampleData();
    } else {
      final messenger = ScaffoldMessenger.of(context);
      final result = await fetchAttributes();
      switch (result) {
        case Success(value: final value):
          _data = value;
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
          _data = {};
          return;
      }
    }
    getAllAttributes(_data!);
    applyMathFunctions(_data!); // Calculate sum and average
    print("GOT DATA");
    // print(_data);
  }

  Map<String, Map<String, String>> getSampleData() {
    var rng = Random();
    Map<String, Map<String, String>> sampleData = {};
    for (var listObjs in kDataSample.values) {
      for (var obj in listObjs) {
        sampleData[obj] = {
          "height": rng.nextInt(100).toString(),
          "weight": "45.5",
          "vendor": "test"
        };
      }
    }
    return sampleData;
  }

  getAllAttributes(Map<String, Map<String, dynamic>> data) {
    for (var obj in widget.selectedObjects) {
      if (data.containsKey(obj)) {
        for (var attr in data[obj]!.keys) {
          _allAttributes.add(attr);
        }
      }
    }
    for (String attr in ["height", "size"]) {
      if (_allAttributes.contains(attr) && !selectedAttrs.contains(attr)) {
        selectedAttrs.add(attr);
      }
    }
    if (selectedAttrs.length < 3) {
      if (!selectedAttrs.contains(_allAttributes.first) &&
          _allAttributes.first != extraColumn) {
        selectedAttrs.add(_allAttributes.first);
      }
    }
    addColumnLabels();
  }

  addColumnLabels() {
    // Column labels
    // First, the objects column
    columnLabels = [
      DataColumn(
          label: const Text(
            "Objects",
            style: TextStyle(fontWeight: FontWeight.w600),
          ),
          onSort: (columnIndex, ascending) {
            setState(() {
              sort = !sort;
              sortColumnIndex = columnIndex;
            });
            onsortColum(columnIndex, ascending);
          })
    ];
    // Then all selected attributes
    for (var attr in selectedAttrs) {
      if (attr != extraColumn) {
        columnLabels.add(DataColumn(
            label: Row(
          children: [
            Text(
              attr,
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
          ],
        )));
      }
    }
    // Finally, add new column
    columnLabels.add(DataColumn(
      label: PopupMenuButton<String>(
        tooltip: AppLocalizations.of(context)!.addColumnTip,
        offset: const Offset(0, -32),
        itemBuilder: (_) => attributesCheckList(selectedAttrs),
        onCanceled: () => print('canceled'),
        icon: const Icon(Icons.add),
      ),
    ));
  }

  applyMathFunctions(Map<String, Map<String, dynamic>> data) {
    List<String> mathFunctions = [sumStr, avgStr];
    for (var func in mathFunctions) {
      data[func] = {};
      for (String attr in _allAttributes) {
        double? sum;
        var count = 0;
        for (var obj in widget.selectedObjects) {
          if (data.containsKey(obj) && data[obj]!.containsKey(attr)) {
            var currentValue = data[obj]![attr]!.toString();
            double? value = double.tryParse(currentValue);
            if (value != null) {
              count++;
              if (sum == null) {
                sum = value;
              } else {
                sum += value;
              }
            }
          }
        }
        if (sum != null) {
          data[func]![attr] = func == sumStr
              ? sum.toStringAsFixed(2)
              : (sum / count).toStringAsFixed(2);
        }
      }
      print(data[func]);
    }
  }

  List<PopupMenuEntry<String>> attributesCheckList(List<String> selectedAttrs) {
    return _allAttributes.map((String key) {
      return PopupMenuItem(
        padding: EdgeInsets.zero,
        height: 0,
        value: key,
        child: StatefulBuilder(builder: (context, localSetState) {
          return CheckboxListTile(
            controlAffinity: ListTileControlAffinity.leading,
            title: Text(key),
            value: selectedAttrs.contains(key),
            dense: true,
            onChanged: (bool? value) {
              setState(() {
                if (value!) {
                  selectedAttrs.add(key);
                } else {
                  selectedAttrs.remove(key);
                }
                addColumnLabels();
              });
              localSetState(() {});
            },
          );
        }),
      );
    }).toList();
  }

  List<PopupMenuEntry<String>> mathFunctionsPopup() {
    final localeMsg = AppLocalizations.of(context)!;
    return <PopupMenuEntry<String>>[
      PopupMenuItem(
        value: sumStr,
        textStyle: Theme.of(context).textTheme.bodyMedium,
        onTap: () {
          setState(() {
            if (!widget.selectedObjects.contains(sumStr)) {
              widget.selectedObjects.insert(0, sumStr);
            } else {
              widget.selectedObjects.remove(sumStr);
            }
          });
        },
        child: Text(localeMsg.showSum),
      ),
      PopupMenuItem(
        value: avgStr,
        textStyle: Theme.of(context).textTheme.bodyMedium,
        onTap: () {
          setState(() {
            if (!widget.selectedObjects.contains(avgStr)) {
              widget.selectedObjects.insert(0, avgStr);
            } else {
              widget.selectedObjects.remove(avgStr);
            }
          });
        },
        child: Text(localeMsg.showAvg),
      ),
    ];
  }

  getCSV() async {
    // Prepare data
    final firstRow = ["Objects", ...selectedAttrs];
    firstRow.remove(extraColumn);
    List<List<String>> rows = [firstRow];
    for (var obj in widget.selectedObjects) {
      List<String> row = [];
      row.add(obj);
      for (String attr in selectedAttrs) {
        if (attr != extraColumn) {
          String value = "-";
          if (_data!.containsKey(obj) && _data![obj]!.containsKey(attr)) {
            value = _data![obj]![attr]!.toString();
          }
          row.add(value);
        }
      }
      rows.add(row);
    }

    // Prepare the file
    String csv = const ListToCsvConverter().convert(rows);
    final bytes = utf8.encode(csv);
    if (kIsWeb) {
      // If web, use html to download csv
      html.AnchorElement(
          href: 'data:application/octet-stream;base64,${base64Encode(bytes)}')
        ..setAttribute("download", "report.csv")
        ..click();
    } else {
      // Save to local filesystem
      var path = (await getApplicationDocumentsDirectory()).path;
      var fileName = '$path/report.csv';
      var file = File(fileName);
      for (var i = 1; await file.exists(); i++) {
        fileName = '$path/report ($i).csv';
        file = File(fileName);
      }
      file.writeAsBytes(bytes, flush: true).then((value) => showSnackBar(
          ScaffoldMessenger.of(context),
          "${AppLocalizations.of(context)!.fileSavedTo} $fileName"));
    }
  }
}

class CustomRow {
  CustomRow(
    this.cells,
  );

  final List<DataCell> cells;

  bool selected = false;
}

class _DataSource extends DataTableSource {
  List<String> selectedAttrs;
  List<String> selectedObjects;
  Map<String, Map<String, dynamic>>? data;

  _DataSource(
      this.context, this.selectedAttrs, this.selectedObjects, this.data) {
    _rows = getChildren();
  }
  final BuildContext context;
  late List<CustomRow> _rows;

  final int _selectedCount = 0;

  @override
  DataRow? getRow(int index) {
    assert(index >= 0);
    if (index >= _rows.length) return null;
    final row = _rows[index];
    return DataRow.byIndex(
      index: index,
      selected: row.selected,
      // onSelectChanged: (value) {
      //   if (row.selected != value) {
      //     _selectedCount += value! ? 1 : -1;
      //     assert(_selectedCount >= 0);
      //     row.selected = value;
      //     notifyListeners();
      //   }
      // },
      cells: row.cells,
    );
  }

  @override
  int get rowCount => _rows.length;

  @override
  bool get isRowCountApproximate => false;

  @override
  int get selectedRowCount => _selectedCount;

  List<CustomRow> getChildren() {
    List<CustomRow> children = [];
    for (var obj in selectedObjects) {
      List<DataCell> row = [];
      row.add(label(obj, fontWeight: FontWeight.w600));
      for (String attr in selectedAttrs) {
        if (attr != extraColumn) {
          String value = "-";
          if (data!.containsKey(obj) && data![obj]!.containsKey(attr)) {
            value = data![obj]![attr]!.toString();
          }
          row.add(label(value));
        }
      }
      // for add column at the end
      if (selectedAttrs.contains(extraColumn)) row.add(label(""));
      children.add(CustomRow(row));
    }
    return children;
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: SelectableText(
          label,
          style: TextStyle(
              fontSize: 14,
              fontWeight: fontWeight,
              color: label.contains('(') ? Colors.green : null),
        ),
      ),
    );
  }
}
