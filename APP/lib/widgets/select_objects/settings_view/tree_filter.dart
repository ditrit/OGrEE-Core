import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/settings_view.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

const lastLevel = 3;

extension StringExtension on String {
  String capitalize() {
    return "${this[0].toUpperCase()}${substring(1).toLowerCase()}";
  }
}

class TreeFilter extends StatefulWidget {
  const TreeFilter({super.key});

  @override
  State<TreeFilter> createState() => TreeFilterState();

  static TreeFilterState? of(BuildContext context) =>
      context.findAncestorStateOfType<TreeFilterState>();
}

class TreeFilterState extends State<TreeFilter> {
  Map<int, List<String>> _filterLevels = {0: [], 1: [], 2: [], 3: []};
  Map<int, List<String>> get filterLevels => _filterLevels;

  Map<String, List<String>> objectsPerCategory = {};
  Map<String, int> enumParams = {};

  @override
  Widget build(BuildContext context) {
    _filterLevels = TreeAppController.of(context).filterLevels;
    // Get which fields to filter and their list of suggestions
    int idx = 0;
    if (TreeAppController.of(context).fetchedCategories["KeysOrder"] != null) {
      for (final String key
          in TreeAppController.of(context).fetchedCategories["KeysOrder"]!) {
        objectsPerCategory[key.capitalize()] =
            TreeAppController.of(context).fetchedCategories[key] ??
                []; // field name
        enumParams[key.capitalize()] = idx; // field name -> id
        idx++;
      }
    }

    return Column(
      children: objectsPerCategory.keys.map((key) {
        // Input enabled only if child of selected filter or if last level
        final enabled = enumParams[key]! > getMaxFilterLevel() ||
            enumParams[key]! == lastLevel;
        List<String> options = objectsPerCategory[key]!;

        // Update suggestions according to last selected level
        if (enabled && !isFilterEmpty(topLevel: lastLevel - 1)) {
          final lastLevelFilters =
              _filterLevels[getMaxFilterLevel(topLevel: lastLevel - 1)]!;
          options = options.where((obj) {
            for (final filter in lastLevelFilters) {
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
          notifyParent: notifyChildSelection,
          showClearFilter: enumParams[key] == 0 ? !isFilterEmpty() : false,
        );
      }).toList(),
    );
  }

  // Callback for child to update parent state
  void notifyChildSelection({bool isClearAll = false}) {
    if (isClearAll) {
      for (final level in _filterLevels.keys) {
        _filterLevels[level] = [];
      }
      TreeAppController.of(context).filterTree("", -1);
    }
    setState(() {});
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
  final Function({bool isClearAll}) notifyParent;
  final bool showClearFilter;

  const AutocompleteFilter({
    super.key,
    required this.enabled,
    required this.param,
    required this.paramLevel,
    required this.options,
    required this.notifyParent,
    required this.showClearFilter,
  });

  @override
  State<AutocompleteFilter> createState() => _AutocompleteFilterState();
}

const Color kDarkBlue = Color(0xff1565c0);

class _AutocompleteFilterState extends State<AutocompleteFilter> {
  List<String> _selectedOptions = []; // overwritten at init by parent ref

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _selectedOptions = TreeFilter.of(context)!.filterLevels[widget.paramLevel]!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (widget.paramLevel == 0)
          Wrap(
            children: [
              SettingsHeader(text: localeMsg.categoryFilters),
              if (widget.showClearFilter)
                OutlinedButton(
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.orange.shade700,
                    backgroundColor: Colors.orange.shade100,
                    padding: const EdgeInsets.all(8),
                    side: const BorderSide(style: BorderStyle.none),
                    shape: const RoundedRectangleBorder(
                      borderRadius: BorderRadius.all(Radius.circular(12)),
                    ),
                    textStyle: const TextStyle(
                      color: kDarkBlue,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  onPressed: () => widget.notifyParent(isClearAll: true),
                  child: Text(localeMsg.clearAllFilters),
                )
              else
                Container(),
            ],
          )
        else
          Container(),
        RawAutocomplete<String>(
          optionsBuilder: (TextEditingValue textEditingValue) {
            return widget.options.where((String option) {
              return option.contains(textEditingValue.text);
            });
          },
          fieldViewBuilder: (
            BuildContext context,
            TextEditingController textEditingController,
            FocusNode focusNode,
            VoidCallback onFieldSubmitted,
          ) {
            return TextFormField(
              controller: textEditingController,
              focusNode: focusNode,
              style: const TextStyle(fontSize: 14),
              decoration: GetFormInputDecoration(
                true,
                widget.param,
                isEnabled: widget.enabled,
              ),
              onFieldSubmitted: (String value) {
                if (widget.options.contains(value)) {
                  setState(() {
                    TreeAppController.of(context)
                        .filterTree(value, widget.paramLevel);
                    widget.notifyParent();
                    // _selectedOptions.add(value);
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
          optionsViewBuilder: (
            BuildContext context,
            AutocompleteOnSelected<String> onSelected,
            Iterable<String> options,
          ) {
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
                          title: Text(
                            option,
                            style: const TextStyle(fontSize: 14),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ),
            );
          },
        ),
        Wrap(
          spacing: 10,
          runSpacing: 10,
          children: getChips(_selectedOptions, context),
        ),
        const SizedBox(height: 4),
      ],
    );
  }

  // One chip per selected filter
  List<Widget> getChips(List<String> nodes, BuildContext context) {
    final List<Widget> chips = [];
    for (final value in nodes) {
      chips.add(
        RawChip(
          onPressed: () {
            TreeAppController.of(context).filterTree(value, widget.paramLevel);
            setState(() {
              // _selectedOptions.removeWhere((opt) => opt == value);
              widget.notifyParent();
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
        ),
      );
    }
    return chips;
  }
}
