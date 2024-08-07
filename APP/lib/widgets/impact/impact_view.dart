import 'dart:convert';
import 'dart:io';

import 'package:csv/csv.dart';
import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/widgets/impact/impact_graph_view.dart';
import 'package:ogree_app/widgets/impact/impact_popup.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:path_provider/path_provider.dart';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:universal_html/html.dart' as html;

class ImpactView extends StatefulWidget {
  String rootId;
  bool? receivedMarkAll;
  ImpactView({required this.rootId, required this.receivedMarkAll});

  @override
  State<ImpactView> createState() => _ImpactViewState();
}

class _ImpactViewState extends State<ImpactView>
    with AutomaticKeepAliveClientMixin {
  Map<String, dynamic> _data = {};
  List<String> selectedCategories = [];
  List<String> selectedPtypes = [];
  List<String> selectedVtypes = [];
  bool isMarkedForMaintenance = false;
  bool? lastReceivedMarkAll;

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    if ('.'.allMatches(widget.rootId).length > 2) {
      // default for racks and under
      selectedPtypes = ["blade"];
      selectedVtypes = ["application", "cluster", "vm"];
    }
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    bool isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

    if (lastReceivedMarkAll != widget.receivedMarkAll) {
      isMarkedForMaintenance = widget.receivedMarkAll!;
      lastReceivedMarkAll = widget.receivedMarkAll;
    }

    return FutureBuilder(
        future: _data.isEmpty ? getData() : null,
        builder: (context, _) {
          if (_data.isEmpty) {
            return SizedBox(
              height: MediaQuery.of(context).size.height > 205
                  ? MediaQuery.of(context).size.height - 220
                  : MediaQuery.of(context).size.height,
              child: const Card(
                margin: EdgeInsets.all(0.1),
                child: Center(child: CircularProgressIndicator()),
              ),
            );
          }
          return objectImpactView(widget.rootId, localeMsg);
        });
  }

  getData() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchObjectImpact(
        widget.rootId, selectedCategories, selectedPtypes, selectedVtypes);
    switch (result) {
      case Success(value: final value):
        _data = value;
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        _data = {};
        return;
    }
    // is it marked for maintenance?
    final alertResult = await fetchAlert(widget.rootId);
    switch (alertResult) {
      case Success():
        isMarkedForMaintenance = true;
      case Failure():
        return;
    }
  }

  objectImpactView(String rootId, AppLocalizations localeMsg) {
    print(widget.receivedMarkAll);
    return Column(
      children: [
        const SizedBox(height: 10),
        Row(
          children: [
            Padding(
              padding: EdgeInsets.only(left: 4),
              child: SizedBox(
                width: 230,
                child: TextButton.icon(
                    onPressed: () => markForMaintenance(localeMsg),
                    label: Text(isMarkedForMaintenance
                        ? localeMsg.markedMaintenance
                        : localeMsg.markMaintenance),
                    icon: isMarkedForMaintenance
                        ? Icon(Icons.check_circle)
                        : Icon(Icons.check_circle_outline)),
              ),
            ),
            Expanded(
              child: Padding(
                padding: EdgeInsets.only(right: 150),
                child: getWidgetSpan(rootId, "target", size: 18),
              ),
            ),
            IconButton(onPressed: () => getCSV(), icon: Icon(Icons.download)),
            Padding(
              padding: EdgeInsets.only(right: 10),
              child: IconButton(
                  onPressed: () => showCustomPopup(
                      context,
                      ImpactOptionsPopup(
                        selectedCategories: selectedCategories,
                        selectedPtypes: selectedPtypes,
                        selectedVtypes: selectedVtypes,
                        parentCallback: changeImpactFilters,
                      ),
                      isDismissible: true),
                  icon: Icon(Icons.edit)),
            ),
          ],
        ),

        Align(
          alignment: Alignment.center,
          child: Padding(
            padding: const EdgeInsets.only(top: 16),
            child: Text(
              localeMsg.impacts,
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.only(top: 16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Column(
                children: [
                  Row(
                    children: [
                      Text(
                        localeMsg.directly.toUpperCase(),
                        style: TextStyle(
                            fontWeight: FontWeight.w900, fontSize: 17),
                      ),
                      Padding(
                        padding: EdgeInsets.only(left: 6),
                        child: Tooltip(
                          message: localeMsg.directTip,
                          verticalOffset: 13,
                          decoration: BoxDecoration(
                            color: Colors.blueAccent,
                            borderRadius: BorderRadius.all(Radius.circular(12)),
                          ),
                          textStyle: TextStyle(
                            fontSize: 13,
                            color: Colors.white,
                          ),
                          padding: EdgeInsets.all(13),
                          child: Icon(Icons.info_outline_rounded,
                              color: Colors.blueAccent),
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 15),
                  ...listImpacted(_data["direct"]),
                ],
              ),
              Column(
                mainAxisAlignment: MainAxisAlignment.start,
                mainAxisSize: MainAxisSize.max,
                children: [
                  Row(
                    children: [
                      Text(
                        localeMsg.indirectly.toUpperCase(),
                        style: TextStyle(
                            fontWeight: FontWeight.w900, fontSize: 17),
                      ),
                      Padding(
                        padding: EdgeInsets.only(left: 6),
                        child: Tooltip(
                          message: localeMsg.indirectTip,
                          verticalOffset: 13,
                          decoration: BoxDecoration(
                            color: Colors.blueAccent,
                            borderRadius: BorderRadius.all(Radius.circular(12)),
                          ),
                          textStyle: TextStyle(
                            fontSize: 13,
                            color: Colors.white,
                          ),
                          padding: EdgeInsets.all(13),
                          child: Icon(Icons.info_outline_rounded,
                              color: Colors.blueAccent),
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 15),
                  ...listImpacted(_data["indirect"]),
                ],
              )
            ],
          ),
        ),
        SizedBox(height: 15),
        Center(
          child: Text(
            localeMsg.graphView,
            style: Theme.of(context).textTheme.headlineMedium,
          ),
        ),
        SizedBox(height: 15),
        // ImpactGraphView("BASIC.A.R1.A01.chT"),
        ImpactGraphView(rootId, _data),
        SizedBox(height: 10),
      ],
    );
  }

  getCSV() async {
    // Prepare data
    List<List<String>> rows = [
      ["target", widget.rootId]
    ];
    for (var type in ["direct", "indirect"]) {
      var direct = (Map<String, dynamic>.from(_data[type])).keys.toList();
      direct.insertAll(0, [type]);
      rows.add(direct);
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

  markForMaintenance(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (isMarkedForMaintenance) {
      // unmark
      var result = await deleteObject(widget.rootId, "alert");
      switch (result) {
        case Success():
          showSnackBar(
            messenger,
            "${widget.rootId} ${localeMsg.isUnmarked}",
          );
          setState(() {
            isMarkedForMaintenance = false;
          });
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    } else {
      var alert = Alert(widget.rootId, "minor",
          "${widget.rootId} ${localeMsg.isMarked}", localeMsg.checkImpact);

      var result = await createAlert(alert);
      switch (result) {
        case Success():
          showSnackBar(messenger, "${widget.rootId} ${localeMsg.successMarked}",
              isSuccess: true);
          setState(() {
            isMarkedForMaintenance = true;
          });
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  getWidgetSpan(String text, String category, {double size = 14}) {
    MaterialColor badgeColor = Colors.blue;
    if (category == "device") {
      badgeColor = Colors.teal;
    } else if (category == "virtual_obj") {
      badgeColor = Colors.deepPurple;
    }
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(
            height: size + 10,
            child: Tooltip(
              message: category,
              verticalOffset: 13,
              child: Badge(
                backgroundColor: badgeColor.shade50,
                label: Text(
                  " $text ",
                  style: TextStyle(
                      fontSize: size,
                      fontWeight: FontWeight.bold,
                      color: badgeColor.shade900),
                ),
              ),
            ),
          ),
          // Text(text),
        ],
      ),
    );
  }

  List<Widget> listImpacted(Map<String, dynamic> objects) {
    List<Widget> listWidgets = [];
    for (var objId in objects.keys) {
      listWidgets.add(getWidgetSpan(objId, objects[objId]["category"]));
    }
    return listWidgets;
  }

  changeImpactFilters(
      List<String> categories, List<String> ptypes, List<String> vtypes) async {
    selectedCategories = categories;
    selectedPtypes = ptypes;
    selectedVtypes = vtypes;
    setState(() {
      _data = {};
    });
    await getData();
  }
}
