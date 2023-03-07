import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'settings_view.dart';

const lastLevel = 3;

extension StringExtension on String {
  String capitalize() {
    return "${this[0].toUpperCase()}${substring(1).toLowerCase()}";
  }
}

class TreeFilter extends StatefulWidget {
  TreeFilter({super.key});

  @override
  State<TreeFilter> createState() => _TreeFilterState();
}

class _TreeFilterState extends State<TreeFilter> {
  final Map<int, List<String>> _filterLevels = {0: [], 1: [], 2: [], 3: []};
  Map<String, List<String>> objectsPerCategory = {};
  Map<String, int> enumParams = {};

  @override
  Widget build(BuildContext context) {
    int idx = 0;
    for (String key
        in AppController.of(context).fetchedCategories["KeysOrder"]!) {
      objectsPerCategory[key.capitalize()] =
          AppController.of(context).fetchedCategories[key]!;
      enumParams[key.capitalize()] = idx;
      idx++;
    }

    // print(objectsPerCategory);
    return Column(
        children: objectsPerCategory.keys.map((key) {
      var enabled = enumParams[key]! > getMaxFilterLevel() ||
          enumParams[key]! == lastLevel;
      List<String> options = objectsPerCategory[key]!;

      // Apply last level filters to current options
      if (enabled && !isFilterEmpty(topLevel: lastLevel - 1)) {
        var lastLevelFilters =
            _filterLevels[getMaxFilterLevel(topLevel: lastLevel - 1)]!;
        options = options.where((obj) {
          for (var filter in lastLevelFilters) {
            if (obj.contains(filter)) return true;
          }
          return false;
        }).toList();
      }

      // Special filter for last level with multiple selection
      if (enumParams[key]! == lastLevel &&
          _filterLevels[lastLevel]!.isNotEmpty) {
        options = options
            .where((obj) => !_filterLevels[lastLevel]!.contains(obj))
            .toList();
      }

      return AutocompleteFilter(
        enabled: enabled,
        param: key,
        paramLevel: enumParams[key]!,
        options: options,
        notifyParent: notifySelection,
        showClearFilter: enumParams[key] == 0 ? !isFilterEmpty() : false,
      );
    }).toList());
  }

  void notifySelection(String param, String filter, bool selected) {
    if (filter == "CLEAR ALL") {
      // setState(() {
      //   for (var value in _filterLevels.values) {
      //     value = [];
      //   }
      // });
    } else {
      setState(() => selected
          ? _filterLevels[enumParams[param]]!.add(filter)
          : _filterLevels[enumParams[param]]!.remove(filter));
    }
  }

  int getMaxFilterLevel({int topLevel = 3}) {
    var testLevel = topLevel;
    while (testLevel >= 0 && _filterLevels[testLevel]!.isEmpty) {
      testLevel--;
    }
    return testLevel;
  }

  bool isFilterEmpty({int topLevel = 3}) {
    for (var i = 0; i <= topLevel; i++) {
      if (_filterLevels[i]!.isNotEmpty) return false;
    }
    return true;
  }
}

Map<String, MaterialColor> ColorChip = {
  "Site": Colors.teal,
  "Building": Colors.lightBlue,
  "Room": Colors.purple,
  "Rack": Colors.indigo,
};

class AutocompleteFilter extends StatefulWidget {
  final bool enabled;
  final String param;
  final int paramLevel;
  final List<String> options;
  final Function(String, String, bool) notifyParent;
  final bool showClearFilter;

  const AutocompleteFilter(
      {super.key,
      required this.enabled,
      required this.param,
      required this.paramLevel,
      required this.options,
      required this.notifyParent,
      required this.showClearFilter});

  @override
  State<AutocompleteFilter> createState() => _AutocompleteFilterState();
}

const Color kDarkBlue = Color(0xff1565c0);

class _AutocompleteFilterState extends State<AutocompleteFilter> {
  List<String> _selectedOptions = [];

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        widget.paramLevel == 0
            ? Wrap(
                children: [
                  SettingsHeader(text: localeMsg.filters),
                  widget.showClearFilter
                      ? OutlinedButton(
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.orange.shade700,
                            backgroundColor: Colors.orange.shade100,
                            padding: const EdgeInsets.all(8),
                            side: const BorderSide(style: BorderStyle.none),
                            shape: const RoundedRectangleBorder(
                              borderRadius:
                                  BorderRadius.all(Radius.circular(12)),
                            ),
                            textStyle: const TextStyle(
                              color: kDarkBlue,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          onPressed: () => widget.notifyParent(
                              widget.param, "CLEAR ALL", true),
                          child: Text(localeMsg.clearAllFilters),
                        )
                      : Container(),
                ],
              )
            : Container(),
        RawAutocomplete<String>(
          optionsBuilder: (TextEditingValue textEditingValue) {
            return widget.options.where((String option) {
              return option.contains(textEditingValue.text);
            });
          },
          fieldViewBuilder: (BuildContext context,
              TextEditingController textEditingController,
              FocusNode focusNode,
              VoidCallback onFieldSubmitted) {
            return TextFormField(
              controller: textEditingController,
              focusNode: focusNode,
              decoration: InputDecoration(
                  enabled: widget.enabled,
                  isDense: true,
                  labelText: widget.param,
                  labelStyle:
                      const TextStyle(fontSize: 14)), // placeholder text
              onFieldSubmitted: (String value) {
                // print("submitted " + value);
                if (widget.options.contains(value)) {
                  setState(() {
                    AppController.of(context)
                        .filterTree(value, widget.paramLevel);
                    widget.notifyParent(widget.param, value, true);
                    _selectedOptions.add(value);
                    textEditingController.clear();
                  });
                }
                onFieldSubmitted();
              },
              onTap: () {
                // force call optionsBuilder for
                // when widgets.options changes
                textEditingController.notifyListeners();
              },
            );
          },
          optionsViewBuilder: (BuildContext context,
              AutocompleteOnSelected<String> onSelected,
              Iterable<String> options) {
            // print("options view builder");
            return Align(
              alignment: Alignment.topLeft,
              child: Material(
                elevation: 4.0,
                child: SizedBox(
                  height: options.length > 2 ? 150.0 : 50.0 * options.length,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(8.0),
                    itemCount: options.length,
                    itemBuilder: (BuildContext context, int index) {
                      final String option = options.elementAt(index);
                      return GestureDetector(
                        onTap: () {
                          onSelected(option);
                        },
                        child: ListTile(
                          title: Text(option,
                              style: const TextStyle(fontSize: 14)),
                        ),
                      );
                    },
                  ),
                ),
              ),
            );
          },
        ),
        const SizedBox(height: 4),
        Wrap(
          spacing: 10,
          runSpacing: 10,
          children: getChips(_selectedOptions, context),
        ),
      ],
    );
  }

  List<Widget> getChips(List<String> nodes, BuildContext context) {
    List<Widget> chips = [];
    nodes.forEach((value) {
      chips.add(RawChip(
        onPressed: () {
          AppController.of(context).filterTree(value, widget.paramLevel);
          setState(() {
            _selectedOptions.removeWhere((opt) => opt == value);
            widget.notifyParent(widget.param, value, false);
          });
        },
        backgroundColor: ColorChip[widget.param]!.shade100,
        side: const BorderSide(style: BorderStyle.none),
        label: Text(
          value,
          style: TextStyle(
            fontSize: 13.5,
            fontFamily: GoogleFonts.inter().fontFamily,
            color: ColorChip[widget.param],
            fontWeight: FontWeight.w600,
          ),
        ),
        avatar: Icon(
          Icons.cancel,
          size: 20,
          color: ColorChip[widget.param],
        ),
      ));
    });
    return chips;
  }
}
