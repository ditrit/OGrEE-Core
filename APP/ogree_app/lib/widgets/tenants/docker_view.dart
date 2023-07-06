import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/tenants/popups/container_logs_popup.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class DockerView extends StatelessWidget {
  final String tName;
  List<DockerContainer>? _dockerInfo;

  DockerView({super.key, required this.tName});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return FutureBuilder(
        future: getData(),
        builder: (context, _) {
          if (_dockerInfo == null) {
            return const Center(child: CircularProgressIndicator());
          } else if (_dockerInfo!.isEmpty) {
            return Text(localeMsg.noDockerInfo);
          } else {
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
                    rowsPerPage: _dockerInfo!.length,
                    columns: getColumns(),
                    source: _DataSource(context, _dockerInfo!),
                  ),
                ));
          }
        });
  }

  getData() async {
    _dockerInfo = await fetchTenantDockerInfo(tName);
  }

  List<DataColumn> getColumns() {
    TextStyle titleStyle = const TextStyle(fontWeight: FontWeight.w600);
    List<DataColumn> columns = [];
    for (var col in [
      "Logs",
      "Name",
      "Last Started",
      "Status",
      "Image",
      "Size",
      "Port(s)"
    ]) {
      columns.add(DataColumn(
          label: Text(
        col,
        style: titleStyle,
      )));
    }
    return columns;
  }
}

class _DataSource extends DataTableSource {
  List<DockerContainer> dockerList;
  _DataSource(this.context, this.dockerList) {
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
    for (var container in dockerList) {
      List<DataCell> row = [];
      row.add(DataCell(Align(
        alignment: Alignment.centerLeft,
        child: CircleAvatar(
          radius: 13,
          child: IconButton(
              splashRadius: 18,
              iconSize: 14,
              padding: const EdgeInsets.all(2),
              onPressed: () => showCustomPopup(
                  context, ContainerLogsPopup(containerName: container.name)),
              icon: const Icon(
                Icons.search,
              )),
        ),
      )));
      row.addAll([
        label(container.name),
        label(container.lastStarted),
        label(container.status),
        label(container.image),
        label(container.size),
        label(container.ports)
      ]);
      children.add(CustomRow(row));
    }
    return children;
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(getDockerText(label));
  }

  getDockerText(String value) {
    if (value.contains("run")) {
      return Row(children: [
        const Icon(Icons.directions_run, color: Colors.green),
        Text(
          value.capitalize(),
          style: const TextStyle(color: Colors.green),
        )
      ]);
    } else if (value.contains("exit")) {
      return Row(children: [
        const Icon(Icons.error_outline, color: Colors.red),
        const SizedBox(width: 2),
        Text(
          value.capitalize(),
          style: const TextStyle(color: Colors.red),
        )
      ]);
    } else {
      return Text(
        value,
        style: TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.w400,
        ),
      );
    }
  }
}
