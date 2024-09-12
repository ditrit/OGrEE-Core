import 'dart:convert';
import 'dart:io';

import 'package:csv/csv.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/widgets/impact/impact_view.dart';
import 'package:path_provider/path_provider.dart';
import 'package:universal_html/html.dart' as html;

class ImpactPage extends StatefulWidget {
  final List<String> selectedObjects;

  const ImpactPage({
    super.key,
    required this.selectedObjects,
  });

  @override
  State<StatefulWidget> createState() => _ImpactPageState();
}

class _ImpactPageState extends State<ImpactPage> {
  List<DataColumn> columnLabels = [];
  bool? shouldMarkAll;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return SizedBox(
      height: MediaQuery.of(context).size.height > 205
          ? MediaQuery.of(context).size.height - 220
          : MediaQuery.of(context).size.height,
      child: Card(
        margin: const EdgeInsets.all(0.1),
        child: ListView.builder(
          itemCount: widget.selectedObjects.length + 1,
          itemBuilder: (BuildContext context, int index) {
            index = index - 1;
            if (index == -1) {
              if (widget.selectedObjects.length == 1) {
                return const SizedBox(height: 6);
              }
              return Padding(
                padding: const EdgeInsets.only(top: 12.0),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Tooltip(
                          message: localeMsg.markAllTip,
                          child: TextButton.icon(
                            onPressed: () => markAll(true, localeMsg),
                            label: Text(localeMsg.markAllMaintenance),
                            icon: const Icon(Icons.check_circle),
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.only(left: 2),
                          child: Tooltip(
                            message: localeMsg.unmarkAllTip,
                            child: TextButton.icon(
                              onPressed: () => markAll(false, localeMsg),
                              label: Text(localeMsg.unmarkAll),
                              icon: const Icon(Icons.check_circle_outline),
                            ),
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.only(left: 2),
                          child: Tooltip(
                            message: localeMsg.downloadAllTip,
                            child: TextButton.icon(
                              onPressed: () => downloadAll(),
                              label: Text(localeMsg.downloadAll),
                              icon: const Icon(Icons.download),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const Divider(
                      thickness: 0.5,
                      indent: 20,
                      endIndent: 25,
                    ),
                  ],
                ),
              );
            }
            final String option = widget.selectedObjects.elementAt(index);
            return ImpactView(
              rootId: option,
              receivedMarkAll: shouldMarkAll,
            );
          },
        ),
      ),
    );
  }

  Future<void> markAll(bool isMark, AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (!isMark) {
      // unmark
      for (final obj in widget.selectedObjects) {
        final result = await deleteObject(obj, "alert");
        switch (result) {
          case Success():
            break;
          case Failure(exception: final exception):
            showSnackBar(messenger, exception.toString(), isError: true);
        }
      }
      showSnackBar(
        messenger,
        localeMsg.allUnmarked,
      );
      setState(() {
        shouldMarkAll = false;
      });
    } else {
      for (final obj in widget.selectedObjects) {
        final alert = Alert(
          obj,
          "minor",
          "$obj ${localeMsg.isMarked}",
          localeMsg.checkImpact,
        );

        final result = await createAlert(alert);
        switch (result) {
          case Success():
            break;
          case Failure(exception: final exception):
            showSnackBar(messenger, exception.toString(), isError: true);
            return;
        }
      }
      showSnackBar(messenger, localeMsg.allMarked, isSuccess: true);
      setState(() {
        shouldMarkAll = true;
      });
    }
  }

  Future<void> downloadAll() async {
    final List<List<String>> rows = [];
    final messenger = ScaffoldMessenger.of(context);
    final localeMsg = AppLocalizations.of(context)!;
    for (final obj in widget.selectedObjects) {
      final List<String> selectedCategories = [];
      List<String> selectedPtypes = [];
      List<String> selectedVtypes = [];
      if ('.'.allMatches(obj).length > 2) {
        // default for racks and under
        selectedPtypes = ["blade"];
        selectedVtypes = ["application", "cluster", "vm"];
      }
      final result = await fetchObjectImpact(
        obj,
        selectedCategories,
        selectedPtypes,
        selectedVtypes,
      );
      switch (result) {
        case Success(value: final value):
          rows.add(["target", obj]);
          for (final type in ["direct", "indirect"]) {
            final direct = Map<String, dynamic>.from(value[type]).keys.toList();
            direct.insertAll(0, [type]);
            rows.add(direct);
          }
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
          return;
      }
    }

    // Prepare the file
    final String csv = const ListToCsvConverter().convert(rows);
    final bytes = utf8.encode(csv);
    if (kIsWeb) {
      // If web, use html to download csv
      html.AnchorElement(
        href: 'data:application/octet-stream;base64,${base64Encode(bytes)}',
      )
        ..setAttribute("download", "impact-report.csv")
        ..click();
    } else {
      // Save to local filesystem
      final path = (await getApplicationDocumentsDirectory()).path;
      var fileName = '$path/impact-report.csv';
      var file = File(fileName);
      for (var i = 1; await file.exists(); i++) {
        fileName = '$path/impact-report ($i).csv';
        file = File(fileName);
      }
      file.writeAsBytes(bytes, flush: true).then(
            (value) => showSnackBar(
              messenger,
              "${localeMsg.fileSavedTo} $fileName",
            ),
          );
    }
  }
}
