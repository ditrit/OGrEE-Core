import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/csv.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/widgets/impact/impact_graph_view.dart';
import 'package:ogree_app/widgets/impact/impact_popup.dart';

class ImpactView extends StatefulWidget {
  String rootId;
  bool? receivedMarkAll;
  ImpactView({super.key, required this.rootId, required this.receivedMarkAll});

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
    super.build(context);
    final localeMsg = AppLocalizations.of(context)!;

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
      },
    );
  }

  Future<void> getData() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchObjectImpact(
      widget.rootId,
      selectedCategories,
      selectedPtypes,
      selectedVtypes,
    );
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

  Column objectImpactView(String rootId, AppLocalizations localeMsg) {
    return Column(
      children: [
        const SizedBox(height: 10),
        Row(
          children: [
            Padding(
              padding: const EdgeInsets.only(left: 4),
              child: SizedBox(
                width: 230,
                child: TextButton.icon(
                  onPressed: () => markForMaintenance(localeMsg),
                  label: Text(
                    isMarkedForMaintenance
                        ? localeMsg.markedMaintenance
                        : localeMsg.markMaintenance,
                  ),
                  icon: isMarkedForMaintenance
                      ? const Icon(Icons.check_circle)
                      : const Icon(Icons.check_circle_outline),
                ),
              ),
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.only(right: 150),
                child: getWidgetSpan(rootId, "target", size: 18),
              ),
            ),
            IconButton(
                onPressed: () => getCSV(), icon: const Icon(Icons.download)),
            Padding(
              padding: const EdgeInsets.only(right: 10),
              child: IconButton(
                onPressed: () => showCustomPopup(
                  context,
                  ImpactOptionsPopup(
                    selectedCategories: selectedCategories,
                    selectedPtypes: selectedPtypes,
                    selectedVtypes: selectedVtypes,
                    parentCallback: changeImpactFilters,
                  ),
                  isDismissible: true,
                ),
                icon: const Icon(Icons.edit),
              ),
            ),
          ],
        ),

        Align(
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
                        style: const TextStyle(
                          fontWeight: FontWeight.w900,
                          fontSize: 17,
                        ),
                      ),
                      Padding(
                        padding: const EdgeInsets.only(left: 6),
                        child: Tooltip(
                          message: localeMsg.directTip,
                          verticalOffset: 13,
                          decoration: const BoxDecoration(
                            color: Colors.blueAccent,
                            borderRadius: BorderRadius.all(Radius.circular(12)),
                          ),
                          textStyle: const TextStyle(
                            fontSize: 13,
                            color: Colors.white,
                          ),
                          padding: const EdgeInsets.all(13),
                          child: const Icon(
                            Icons.info_outline_rounded,
                            color: Colors.blueAccent,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 15),
                  ...listImpacted(_data["direct"]),
                ],
              ),
              Column(
                children: [
                  Row(
                    children: [
                      Text(
                        localeMsg.indirectly.toUpperCase(),
                        style: const TextStyle(
                          fontWeight: FontWeight.w900,
                          fontSize: 17,
                        ),
                      ),
                      Padding(
                        padding: const EdgeInsets.only(left: 6),
                        child: Tooltip(
                          message: localeMsg.indirectTip,
                          verticalOffset: 13,
                          decoration: const BoxDecoration(
                            color: Colors.blueAccent,
                            borderRadius: BorderRadius.all(Radius.circular(12)),
                          ),
                          textStyle: const TextStyle(
                            fontSize: 13,
                            color: Colors.white,
                          ),
                          padding: const EdgeInsets.all(13),
                          child: const Icon(
                            Icons.info_outline_rounded,
                            color: Colors.blueAccent,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 15),
                  ...listImpacted(_data["indirect"]),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 15),
        Center(
          child: Text(
            localeMsg.graphView,
            style: Theme.of(context).textTheme.headlineMedium,
          ),
        ),
        const SizedBox(height: 15),
        // ImpactGraphView("BASIC.A.R1.A01.chT"),
        ImpactGraphView(rootId, _data),
        const SizedBox(height: 10),
      ],
    );
  }

  getCSV() async {
    // Prepare data
    final List<List<String>> rows = [
      ["target", widget.rootId],
    ];
    for (final type in ["direct", "indirect"]) {
      final direct = Map<String, dynamic>.from(_data[type]).keys.toList();
      direct.insertAll(0, [type]);
      rows.add(direct);
    }

    // Save the file
    await saveCSV("impact-report", rows, context);
  }

  markForMaintenance(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (isMarkedForMaintenance) {
      // unmark
      final result = await deleteObject(widget.rootId, "alert");
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
      final alert = Alert(
        widget.rootId,
        "minor",
        "${widget.rootId} ${localeMsg.isMarked}",
        localeMsg.checkImpact,
      );

      final result = await createAlert(alert);
      switch (result) {
        case Success():
          showSnackBar(
            messenger,
            "${widget.rootId} ${localeMsg.successMarked}",
            isSuccess: true,
          );
          setState(() {
            isMarkedForMaintenance = true;
          });
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  Padding getWidgetSpan(String text, String category, {double size = 14}) {
    MaterialColor badgeColor = Colors.blue;
    if (category == "device") {
      badgeColor = Colors.teal;
    } else if (category == "virtual_obj") {
      badgeColor = Colors.deepPurple;
    }
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.0),
      child: Row(
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
                    color: badgeColor.shade900,
                  ),
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
    final List<Widget> listWidgets = [];
    for (final objId in objects.keys) {
      listWidgets.add(getWidgetSpan(objId, objects[objId]["category"]));
    }
    return listWidgets;
  }

  changeImpactFilters(
    List<String> categories,
    List<String> ptypes,
    List<String> vtypes,
  ) async {
    selectedCategories = categories;
    selectedPtypes = ptypes;
    selectedVtypes = vtypes;
    setState(() {
      _data = {};
    });
    await getData();
  }
}
