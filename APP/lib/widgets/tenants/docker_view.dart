import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/tenants/popups/backup_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/container_logs_popup.dart';

class DockerView extends StatelessWidget {
  final String tName;
  List<DockerContainer>? _dockerInfo;

  DockerView({super.key, required this.tName});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return FutureBuilder(
      future: getData(context),
      builder: (context, _) {
        if (_dockerInfo == null) {
          return const Center(child: CircularProgressIndicator());
        } else if (_dockerInfo!.isEmpty) {
          return Text(localeMsg.noDockerInfo);
        } else {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Theme(
                data: Theme.of(context).copyWith(
                  cardTheme: const CardTheme(
                    elevation: 0,
                    surfaceTintColor: Colors.white,
                    color: Colors.white,
                  ),
                ),
                child: SingleChildScrollView(
                  padding: const EdgeInsets.only(right: 16),
                  child: PaginatedDataTable(
                    horizontalMargin: 15,
                    columnSpacing: 30,
                    showCheckboxColumn: false,
                    rowsPerPage: _dockerInfo!.length,
                    columns: getColumns(localeMsg),
                    source: _DataSource(context, _dockerInfo!),
                  ),
                ),
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: Padding(
                  padding: const EdgeInsets.only(bottom: 20, right: 20),
                  child: ElevatedButton.icon(
                    onPressed: () => showCustomPopup(
                      context,
                      BackupPopup(tenantName: tName),
                      isDismissible: true,
                    ),
                    icon: const Icon(Icons.history),
                    label: Text(localeMsg.backup.capitalize()),
                  ),
                ),
              ),
            ],
          );
        }
      },
    );
  }

  getData(BuildContext context) async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchTenantDockerInfo(tName);
    switch (result) {
      case Success(value: final value):
        _dockerInfo = value;
      case Failure(exception: final exception):
        showSnackBar(
          messenger,
          exception.toString(),
          isError: true,
        );
        _dockerInfo = [];
    }
  }

  List<DataColumn> getColumns(AppLocalizations localeMsg) {
    const TextStyle titleStyle = TextStyle(fontWeight: FontWeight.w600);
    final List<DataColumn> columns = [];
    for (final col in [
      "Logs",
      localeMsg.name,
      localeMsg.lastStarted,
      "Status",
      "Image",
      localeMsg.size,
      "Port(s)",
    ]) {
      columns.add(
        DataColumn(
          label: Text(
            col,
            style: titleStyle,
          ),
        ),
      );
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
    final List<CustomRow> children = [];
    for (final container in dockerList) {
      final List<DataCell> row = [];
      row.add(
        DataCell(
          Align(
            alignment: Alignment.centerLeft,
            child: CircleAvatar(
              radius: 13,
              child: IconButton(
                splashRadius: 18,
                iconSize: 14,
                padding: const EdgeInsets.all(2),
                onPressed: () => showCustomPopup(
                  context,
                  ContainerLogsPopup(containerName: container.name),
                ),
                icon: const Icon(
                  Icons.search,
                ),
              ),
            ),
          ),
        ),
      );
      row.addAll([
        label(container.name),
        label(container.lastStarted),
        label(container.status),
        label(container.image),
        label(container.size),
        label(container.ports),
      ]);
      children.add(CustomRow(row));
    }
    return children;
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(getDockerText(label));
  }

  Widget getDockerText(String value) {
    if (value.contains("run")) {
      return Row(
        children: [
          const Icon(Icons.directions_run, color: Colors.green),
          Text(
            value.capitalize(),
            style: const TextStyle(color: Colors.green),
          ),
        ],
      );
    } else if (value.contains("exit")) {
      return Row(
        children: [
          const Icon(Icons.error_outline, color: Colors.red),
          const SizedBox(width: 2),
          Text(
            value.capitalize(),
            style: const TextStyle(color: Colors.red),
          ),
        ],
      );
    } else {
      return Text(
        value,
        style: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.w400,
        ),
      );
    }
  }
}
