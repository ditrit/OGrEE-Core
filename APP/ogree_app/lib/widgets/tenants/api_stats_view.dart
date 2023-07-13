import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/results_page.dart';

// Define a stateful widget that displays API usage statistics for a given tenant
class ApiStatsView extends StatefulWidget {
  ApiStatsView({
    Key? key,
  }) : super(key: key);

  @override
  State<ApiStatsView> createState() => _ApiStatsViewState();
}

// Define the state for the ApiStatsView widget
class _ApiStatsViewState extends State<ApiStatsView> {
  Map<String, dynamic>? _tenantStats;
  TextStyle titleStyle = const TextStyle(fontWeight: FontWeight.w600);

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return FutureBuilder(
        // Async method that fetches the tenant's API usage statistics
        future: getTenantStats(),
        builder: (context, _) {
          // If the statistics data is still being fetched, show a loading indicator
          if (_tenantStats == null) {
            return const Center(child: CircularProgressIndicator());
          }
          // If the statistics data is available and not empty, display it
          else if (_tenantStats!.isNotEmpty) {
            return Theme(
                data: ThemeData(
                  cardTheme: const CardTheme(
                      elevation: 0,
                      surfaceTintColor: Colors.white,
                      color: Colors.white),
                ),
                child: SingleChildScrollView(
                  padding: const EdgeInsets.only(right: 16, top: 0),
                  child: PaginatedDataTable(
                    horizontalMargin: 15,
                    columnSpacing: 30,
                    showCheckboxColumn: false,
                    rowsPerPage: _tenantStats!.length,
                    columns: [
                      DataColumn(
                          label: Text(
                        localeMsg.parameter,
                        style: titleStyle,
                      )),
                      DataColumn(
                          label: Text(
                        localeMsg.value,
                        style: titleStyle,
                      ))
                    ],
                    source: _DataSource(context, _tenantStats!),
                  ),
                ));
          }
          // If the statistics data is empty, display a message
          else {
            return Text(localeMsg.noProjects);
          }
        });
  }

  getTenantStats() async {
    // Fetch the statistics data from the tenant's API backend
    Result result = await fetchTenantStats();
    switch (result) {
      case Success(value: final value):
        _tenantStats = value;
      case Failure(exception: final exception):
        showSnackBar(context, exception.toString());
    }

    // Fetch additional version information about the tenant's API
    result = await fetchTenantApiVersion();
    switch (result) {
      case Success(value: final value):
        Map<String, dynamic> versionStats = value;
        for (var key in versionStats.keys) {
          if (key.contains("Build")) {
            _tenantStats!["API$key"] = versionStats[key];
          } else {
            _tenantStats![key] = versionStats[key];
          }
        }
      case Failure(exception: final exception):
        showSnackBar(context, exception.toString());
    }
  }
}

class _DataSource extends DataTableSource {
  Map<String, dynamic> stats;
  _DataSource(this.context, this.stats) {
    _rows = getChildren();
  }
  final BuildContext context;
  late List<CustomRow> _rows;

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
          // notifyListeners();
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

  List<CustomRow> getChildren() {
    List<CustomRow> children = [];
    for (var key in stats.keys) {
      List<DataCell> row = [label(key), label(stats[key].toString())];
      children.add(CustomRow(row));
    }
    return children;
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(Text(
      label,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w400,
      ),
    ));
  }
}
