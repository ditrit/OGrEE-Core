import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/select_objects/object_popup.dart';

class ImpactOptionsPopup extends StatefulWidget {
  List<String> selectedCategories;
  List<String> selectedPtypes;
  List<String> selectedVtypes;
  final Function(
    List<String> selectedCategories,
    List<String> selectedPtypes,
    List<String> selectedVtypes,
  ) parentCallback;
  ImpactOptionsPopup({
    super.key,
    required this.selectedCategories,
    required this.selectedPtypes,
    required this.selectedVtypes,
    required this.parentCallback,
  });

  @override
  State<ImpactOptionsPopup> createState() => _ImpactOptionsPopupState();
}

class _ImpactOptionsPopupState extends State<ImpactOptionsPopup> {
  List<String> ptypes = ["blade", "chassis", "disk", "processor"];
  List<String> vtypes = ["application", "cluster", "storage", "vm"];
  late List<String> selectedCategories;
  late List<String> selectedPtypes;
  late List<String> selectedVtypes;

  @override
  void initState() {
    selectedCategories = widget.selectedCategories;
    selectedPtypes = widget.selectedPtypes;
    selectedVtypes = widget.selectedVtypes;
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        height: 330,
        width: 680,
        constraints: const BoxConstraints(maxHeight: 430),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: ScaffoldMessenger(
            child: Builder(
              builder: (context) => Scaffold(
                backgroundColor: Colors.white,
                body: Column(
                  children: [
                    Center(
                      child: Text(
                        localeMsg.indirectOptions,
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                    ),
                    const SizedBox(height: 15),
                    const SizedBox(height: 10),
                    SizedBox(
                      height: 200,
                      child: Wrap(
                        children: [
                          Column(
                            children: [
                              const Text(
                                "Category",
                              ),
                              SizedBox(
                                height: 200,
                                width: 200,
                                child: ListView.builder(
                                  itemCount: PhyCategories.values.length,
                                  itemBuilder:
                                      (BuildContext context, int index) {
                                    final name =
                                        PhyCategories.values[index].name;
                                    return CheckboxListTile(
                                      controlAffinity:
                                          ListTileControlAffinity.leading,
                                      dense: true,
                                      value: selectedCategories.contains(
                                        PhyCategories.values[index].name,
                                      ),
                                      onChanged: (bool? selected) {
                                        setState(() {
                                          if (selectedCategories.contains(
                                            PhyCategories.values[index].name,
                                          )) {
                                            selectedCategories.remove(
                                              PhyCategories.values[index].name,
                                            );
                                          } else {
                                            selectedCategories.add(
                                              PhyCategories.values[index].name,
                                            );
                                          }
                                        });
                                      },
                                      title: Text(name),
                                    );
                                  },
                                ),
                              ),
                            ],
                          ),
                          Column(
                            children: [
                              const Text("Physical Type"),
                              SizedBox(
                                height: 200,
                                width: 200,
                                child: ListView.builder(
                                  itemCount: ptypes.length,
                                  itemBuilder:
                                      (BuildContext context, int index) {
                                    return CheckboxListTile(
                                      controlAffinity:
                                          ListTileControlAffinity.leading,
                                      dense: true,
                                      value: selectedPtypes
                                          .contains(ptypes[index]),
                                      onChanged: (bool? selected) {
                                        if (selectedPtypes
                                            .contains(ptypes[index])) {
                                          selectedPtypes.remove(ptypes[index]);
                                        } else {
                                          selectedPtypes.add(ptypes[index]);
                                        }
                                        setState(() {});
                                      },
                                      title: Text(ptypes[index]),
                                    );
                                  },
                                ),
                              ),
                            ],
                          ),
                          Column(
                            children: [
                              const Text("Virtual Type"),
                              SizedBox(
                                height: 200,
                                width: 200,
                                child: ListView.builder(
                                  itemCount: vtypes.length,
                                  itemBuilder:
                                      (BuildContext context, int index) {
                                    return CheckboxListTile(
                                      controlAffinity:
                                          ListTileControlAffinity.leading,
                                      dense: true,
                                      value: selectedVtypes
                                          .contains(vtypes[index]),
                                      onChanged: (bool? selected) {
                                        if (selectedVtypes
                                            .contains(vtypes[index])) {
                                          selectedVtypes.remove(vtypes[index]);
                                        } else {
                                          selectedVtypes.add(vtypes[index]);
                                        }
                                        setState(() {});
                                      },
                                      title: Text(vtypes[index]),
                                    );
                                  },
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        ElevatedButton.icon(
                          onPressed: () {
                            widget.parentCallback(
                              selectedCategories,
                              selectedPtypes,
                              selectedVtypes,
                            );
                            Navigator.of(context).pop();
                          },
                          label: const Text("OK"),
                          icon: const Icon(Icons.thumb_up, size: 16),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
