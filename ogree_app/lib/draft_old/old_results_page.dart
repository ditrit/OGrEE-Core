import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api.dart';

class OldResultsPage extends StatefulWidget {
  List<String> selectedAttrs;
  List<String> selectedObjects;

  OldResultsPage(
      {super.key, required this.selectedAttrs, required this.selectedObjects});

  @override
  State<StatefulWidget> createState() => _OldResultsPageState();
}

class _OldResultsPageState extends State<OldResultsPage> {
  Map<String, Map<String, String>>? _data;
  List<Widget> labelRow = [];

  bool sort = true;
  // List<Data>? filterData;

  onsortColum(int columnIndex, bool ascending) {
    if (columnIndex == 0) {
      if (ascending) {
        print("HI");
        // filterData!.sort((a, b) => a.name!.compareTo(b.name!));
      } else {
        print("HELLO");
        // filterData!.sort((a, b) => b.name!.compareTo(a.name!));
      }
    }
  }

  Widget label(String label, {bool isBold = true}) {
    return Center(
      // key: Key(label),
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Text(
          label,
          style: TextStyle(
              fontSize: 14,
              fontWeight: isBold ? FontWeight.w600 : FontWeight.w400),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    print("RESULT");
    labelRow = [label("Objects")];
    print(widget.selectedAttrs);
    for (var attr in widget.selectedAttrs) {
      labelRow.add(label(attr));
    }
    return FutureBuilder(
        future: getData(),
        builder: (context, _) {
          if (_data != null) {
            return Column(
              children: [
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Votre rapport',
                        style: GoogleFonts.inter(
                          fontSize: 23,
                          color: Colors.black,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      Icon(Icons.file_download_outlined)
                    ],
                  ),
                ),
                SizedBox(
                  height: 20,
                ),
                SizedBox(
                  height: MediaQuery.of(context).size.height - 260,
                  child: Card(
                    margin: const EdgeInsets.all(0.1),
                    child: Padding(
                      padding: const EdgeInsets.all(8.0),
                      child: SingleChildScrollView(
                        child:
                            // PaginatedDataTable(
                            //   actions: [Icon(Icons.file_download_outlined)],
                            //   header: Text('Votre rapport'),
                            //   rowsPerPage: 4,
                            //   sortColumnIndex: 0,
                            //   sortAscending: sort,
                            //   columns: labelRow,
                            //   source: _DataSource(context, widget.selectedAttrs,
                            //       widget.selectedObjects, _data),
                            // ),
                            Table(
                                border: TableBorder.all(
                                    width: 0.1,
                                    color: Colors.grey.shade400,
                                    borderRadius: BorderRadius.circular(15)),
                                defaultVerticalAlignment:
                                    TableCellVerticalAlignment.middle,
                                children: getChildren()),
                      ),
                    ),
                  ),
                ),
              ],
            );
          } else {
            return const Center(child: CircularProgressIndicator());
          }
        });
  }

  List<TableRow> getChildren() {
    List<TableRow> children = [];
    children.add(TableRow(children: labelRow));
    print(labelRow.length);
    for (var obj in widget.selectedObjects) {
      List<Widget> row = [];
      row.add(label(obj));
      // print(obj);
      for (String attr in widget.selectedAttrs) {
        // print(attr);
        String value = "-";
        if (_data!.containsKey(obj) && _data![obj]!.containsKey(attr)) {
          value = _data![obj]![attr]!;
        }
        row.add(label(value, isBold: false));
      }
      children.add(TableRow(children: row));
    }
    return children;
  }

  getData() async {
    print("GET DATA");
    _data = await fetchAttributes();
    print("GOT DATA");
    print(_data);
  }
}

class _Row {
  _Row(
    this.cells,
  );

  // final String valueA;
  final List<DataCell> cells;
  // final String valueB;
  // final String valueC;
  // final int valueD;

  bool selected = false;
}

class _DataSource extends DataTableSource {
  List<String> selectedAttrs;
  List<String> selectedObjects;
  Map<String, Map<String, String>>? data;

  _DataSource(
      this.context, this.selectedAttrs, this.selectedObjects, this.data) {
    _rows = getChildren();
    // _rows = <_Row>[
    //   _Row('Cell A1', 'CellB1', 'CellC1', 1),
    //   _Row('Cell A2', 'CellB2', 'CellC2', 2),
    //   _Row('Cell A3', 'CellB3', 'CellC3', 3),
    //   _Row('Cell A4', 'CellB4', 'CellC4', 4),
    // ];
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
    // children.add(TableRow(children: labelRow));
    // print(labelRow.length);
    for (var obj in selectedObjects) {
      List<DataCell> row = [];
      row.add(label(obj));
      // print(obj);
      for (String attr in selectedAttrs) {
        // print(attr);
        String value = "-";
        if (data!.containsKey(obj) && data![obj]!.containsKey(attr)) {
          value = data![obj]![attr]!;
        }
        row.add(label(value, isBold: false));
      }
      children.add(_Row(row));
    }
    return children;
  }

  DataCell label(String label, {bool isBold = true}) {
    return DataCell(
      // key: Key(label),
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: Text(
          label,
          style: TextStyle(
              fontSize: 14,
              fontWeight: isBold ? FontWeight.w600 : FontWeight.w400),
        ),
      ),
    );
  }
}
