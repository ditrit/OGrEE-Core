import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/tenants/popups/container_logs_popup.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class DockerView extends StatelessWidget {
  final String tenantName;
  List<Map<String, String>>? _dockerInfo;

  DockerView({super.key, required this.tenantName});

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
            return Padding(
              padding: const EdgeInsets.all(16.0),
              child: Table(columnWidths: const <int, TableColumnWidth>{
                6: FixedColumnWidth(64),
              }, children: getDockerInfoRows(context)),
            );
          }
        });
  }

  getData() async {
    _dockerInfo = await fetchTenantDockerInfo(tenantName);
  }

  List<TableRow> getDockerInfoRows(BuildContext context) {
    List<TableRow> rows = [];
    rows.add(TableRow(
        children: [
      "Name",
      "Last Started",
      "Status",
      "Image",
      "Size",
      "Port(s)",
      "Logs"
    ]
            .map((e) => Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text(
                    e,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ))
            .toList()));
    for (var container in _dockerInfo!) {
      print(container);
      List<Widget> row = container.values
          .map((e) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 8.0),
                child: getDockerText(e),
              ))
          .toList();
      row.add(Padding(
        padding: const EdgeInsets.symmetric(vertical: 8.0),
        child: Align(
          alignment: Alignment.centerLeft,
          child: CircleAvatar(
            radius: 13,
            child: IconButton(
                splashRadius: 18,
                iconSize: 14,
                padding: const EdgeInsets.all(2),
                onPressed: () => showCustomPopup(context,
                    ContainerLogsPopup(containerName: container["Names"]!)),
                icon: const Icon(
                  Icons.search,
                )),
          ),
        ),
      ));
      rows.add(TableRow(children: row));
    }
    return rows;
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
      return Text(value);
    }
  }
}
