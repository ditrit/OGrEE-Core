import 'dart:convert';
import 'dart:io';
import 'package:csv/csv.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/impact/impact_view.dart';
import 'package:path_provider/path_provider.dart';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:universal_html/html.dart' as html;

class ImpactPage extends StatefulWidget {
  List<String> selectedObjects;

  ImpactPage({
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
    bool isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

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
                return SizedBox(height: 6);
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
                              icon: Icon(Icons.check_circle)),
                        ),
                        Padding(
                          padding: EdgeInsets.only(left: 2),
                          child: Tooltip(
                            message: localeMsg.unmarkAllTip,
                            child: TextButton.icon(
                                onPressed: () => markAll(false, localeMsg),
                                label: Text(localeMsg.unmarkAll),
                                icon: Icon(Icons.check_circle_outline)),
                          ),
                        ),
                        Padding(
                          padding: EdgeInsets.only(left: 2),
                          child: Tooltip(
                            message: localeMsg.downloadAllTip,
                            child: TextButton.icon(
                                onPressed: () => downloadAll(),
                                label: Text(localeMsg.downloadAll),
                                icon: Icon(Icons.download)),
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

  markAll(bool isMark, AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (!isMark) {
      // unmark
      for (var obj in widget.selectedObjects) {
        var result = await deleteObject(obj, "alert");
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
      for (var obj in widget.selectedObjects) {
        var alert = Alert(
            obj, "minor", "$obj ${localeMsg.isMarked}", localeMsg.checkImpact);

        var result = await createAlert(alert);
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

  downloadAll() async {
    List<List<String>> rows = [];
    final messenger = ScaffoldMessenger.of(context);
    for (var obj in widget.selectedObjects) {
      List<String> selectedCategories = [];
      List<String> selectedPtypes = [];
      List<String> selectedVtypes = [];
      if ('.'.allMatches(obj).length > 2) {
        // default for racks and under
        selectedPtypes = ["blade"];
        selectedVtypes = ["application", "cluster", "vm"];
      }
      final result = await fetchObjectImpact(
          obj, selectedCategories, selectedPtypes, selectedVtypes);
      switch (result) {
        case Success(value: final value):
          rows.add(["target", obj]);
          for (var type in ["direct", "indirect"]) {
            var direct = (Map<String, dynamic>.from(value[type])).keys.toList();
            direct.insertAll(0, [type]);
            rows.add(direct);
          }
          break;
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
          return;
      }
    }

    // Prepare the file
    String csv = const ListToCsvConverter().convert(rows);
    final bytes = utf8.encode(csv);
    if (kIsWeb) {
      // If web, use html to download csv
      html.AnchorElement(
          href: 'data:application/octet-stream;base64,${base64Encode(bytes)}')
        ..setAttribute("download", "impact-report.csv")
        ..click();
    } else {
      // Save to local filesystem
      var path = (await getApplicationDocumentsDirectory()).path;
      var fileName = '$path/impact-report.csv';
      var file = File(fileName);
      for (var i = 1; await file.exists(); i++) {
        fileName = '$path/impact-report ($i).csv';
        file = File(fileName);
      }
      file.writeAsBytes(bytes, flush: true).then((value) => showSnackBar(
          ScaffoldMessenger.of(context),
          "${AppLocalizations.of(context)!.fileSavedTo} $fileName"));
    }
  }
}
