import 'package:flutter/material.dart';
import 'package:ogree_app/common/api.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

const String extraColumn = "Add new column";
const String sumStr = "Somme()";
const String avgStr = "Moyenne()";

class ResultsPage extends StatefulWidget {
  List<String> selectedAttrs;
  List<String> selectedObjects;

  ResultsPage(
      {super.key, required this.selectedAttrs, required this.selectedObjects});

  @override
  State<StatefulWidget> createState() => _ResultsPageState();
}

class _ResultsPageState extends State<ResultsPage> {
  Set<String> _allAttributes = {};
  Map<String, Map<String, String>>? _data;
  List<DataColumn> columnLabels = [];

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
    if (!widget.selectedAttrs.contains(extraColumn))
      widget.selectedAttrs.add(extraColumn);
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    print("RESULT");
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
    for (var attr in widget.selectedAttrs) {
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
        tooltip: localeMsg.addColumnTip,
        offset: const Offset(0, -32),
        itemBuilder: (_) => attributesCheckList(widget.selectedAttrs),
        onCanceled: () => print('canceleeed'),
        icon: const Icon(Icons.add),
      ),
    ));

    return FutureBuilder(
        future: _data == null ? getData() : null,
        builder: (context, _) {
          if (_data != null) {
            return SizedBox(
              height: MediaQuery.of(context).size.height - 280,
              child: SingleChildScrollView(
                padding: EdgeInsets.zero,
                child: PaginatedDataTable(
                  header: Text(
                    localeMsg.yourReport,
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                  actions: [
                    const Padding(
                      padding: EdgeInsets.only(right: 8.0),
                      child: Icon(Icons.file_download_outlined),
                    ),
                    PopupMenuButton<String>(
                      tooltip: localeMsg.selectionOptions,
                      offset: const Offset(0, -32),
                      itemBuilder: (_) =>
                          attributesCheckList(widget.selectedAttrs),
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
                  source: _DataSource(context, widget.selectedAttrs,
                      widget.selectedObjects, _data),
                ),
              ),
            );
          } else {
            return const Center(child: CircularProgressIndicator());
          }
        });
  }

  getData() async {
    print("GET DATA");
    _data = await fetchAttributes();
    getAllAttributes(_data!);
    applyMathFunctions(_data!); // Calculate sum and average
    print("GOT DATA");
    print(_data);
  }

  getAllAttributes(Map<String, Map<String, String>> data) {
    for (var obj in widget.selectedObjects) {
      if (data.containsKey(obj)) {
        for (var attr in data[obj]!.keys) {
          _allAttributes.add(attr);
        }
      }
    }
  }

  applyMathFunctions(Map<String, Map<String, String>> data) {
    List<String> mathFunctions = [sumStr, avgStr];
    for (var func in mathFunctions) {
      data[func] = {};
      for (String attr in _allAttributes) {
        double? sum;
        var count = 0;
        for (var obj in widget.selectedObjects) {
          if (data.containsKey(obj) && data[obj]!.containsKey(attr)) {
            double? value = double.tryParse(data[obj]![attr]!);
            if (value != null) {
              count++;
              if (sum == null)
                sum = value;
              else
                sum += value;
            }
          }
        }
        if (sum != null) {
          data[func]![attr] = func == sumStr
              ? sum.toStringAsFixed(3)
              : (sum / count).toStringAsFixed(3);
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
        child: StatefulBuilder(builder: (context, _setState) {
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
              });
              _setState(() {});
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
}

class _Row {
  _Row(
    this.cells,
  );

  final List<DataCell> cells;

  bool selected = false;
}

class _DataSource extends DataTableSource {
  List<String> selectedAttrs;
  List<String> selectedObjects;
  Map<String, Map<String, String>>? data;

  _DataSource(
      this.context, this.selectedAttrs, this.selectedObjects, this.data) {
    _rows = getChildren();
  }
  final BuildContext context;
  late List<_Row> _rows;

  int _selectedCount = 0;

  @override
  DataRow? getRow(int index) {
    assert(index >= 0);
    if (index >= _rows.length) return null;
    final row = _rows[index];
    return DataRow.byIndex(
      index: index,
      selected: row.selected,
      onSelectChanged: (value) {
        if (row.selected != value) {
          _selectedCount += value! ? 1 : -1;
          assert(_selectedCount >= 0);
          row.selected = value;
          notifyListeners();
        }
      },
      cells: row.cells,
    );
  }

  @override
  int get rowCount => _rows.length;

  @override
  bool get isRowCountApproximate => false;

  @override
  int get selectedRowCount => _selectedCount;

  List<_Row> getChildren() {
    List<_Row> children = [];
    for (var obj in selectedObjects) {
      List<DataCell> row = [];
      row.add(label(obj, fontWeight: FontWeight.w500));
      for (String attr in selectedAttrs) {
        if (attr != extraColumn) {
          String value = "-";
          if (data!.containsKey(obj) && data![obj]!.containsKey(attr)) {
            value = data![obj]![attr]!;
          }
          row.add(label(value));
        }
      }
      // for add column at the end
      if (selectedAttrs.contains(extraColumn)) row.add(label(""));
      children.add(_Row(row));
    }
    return children;
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: Text(
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
