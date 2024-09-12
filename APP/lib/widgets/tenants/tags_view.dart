// ignore_for_file: constant_identifier_names

import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tag.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/common/delete_dialog_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/tags_popup.dart';

enum TagSearchFields { Description, Slug, Color }

class TagsView extends StatefulWidget {
  const TagsView({super.key});
  @override
  State<TagsView> createState() => _TagsViewState();
}

class _TagsViewState extends State<TagsView> {
  List<Tag>? _tags;
  bool _loadTags = true;
  List<Tag> selectedTags = [];
  List<Tag>? _filterTags;
  bool sort = true;
  TagSearchFields _searchField = TagSearchFields.Slug;

  onsortColum(int columnIndex, bool ascending) {
    if (columnIndex == 1) {
      if (ascending) {
        _tags!.sort((a, b) => a.slug.compareTo(b.slug));
      } else {
        _tags!.sort((a, b) => b.slug.compareTo(a.slug));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return FutureBuilder(
      future: _loadTags ? getTags() : null,
      builder: (context, _) {
        if (_tags == null) {
          return const Center(child: CircularProgressIndicator());
        }
        return Theme(
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
              sortColumnIndex: 1,
              sortAscending: sort,
              checkboxHorizontalMargin: 0,
              header: Wrap(
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  SizedBox(
                    height: isSmallDisplay ? 30 : 35,
                    width: isSmallDisplay ? 115 : 145,
                    child: DropdownButtonFormField<TagSearchFields>(
                      isExpanded: true,
                      borderRadius: BorderRadius.circular(12.0),
                      decoration: GetFormInputDecoration(
                        isSmallDisplay,
                        null,
                        icon: Icons.search_rounded,
                        contentPadding: isSmallDisplay
                            ? const EdgeInsets.only(
                                bottom: 15,
                                left: 12,
                                right: 5,
                              )
                            : const EdgeInsets.only(
                                top: 3.0,
                                bottom: 12.0,
                                left: 20.0,
                                right: 14.0,
                              ),
                      ),
                      value: _searchField,
                      items: TagSearchFields.values
                          .map<DropdownMenuItem<TagSearchFields>>(
                              (TagSearchFields value) {
                        return DropdownMenuItem<TagSearchFields>(
                          value: value,
                          child: Text(
                            value.name,
                            overflow: TextOverflow.ellipsis,
                          ),
                        );
                      }).toList(),
                      onChanged: (TagSearchFields? value) {
                        setState(() {
                          _searchField = value!;
                        });
                      },
                    ),
                  ),
                  const SizedBox(width: 8),
                  SizedBox(
                    width: 150,
                    child: TextFormField(
                      textAlignVertical: TextAlignVertical.center,
                      onChanged: (value) {
                        setState(() {
                          _tags = searchTags(value);
                        });
                      },
                      decoration: InputDecoration(
                        border: InputBorder.none,
                        isDense: true,
                        label: isSmallDisplay ? null : Text(localeMsg.search),
                        prefixIcon: isSmallDisplay
                            ? const Icon(Icons.search_rounded)
                            : null,
                      ),
                    ),
                  ),
                ],
              ),
              actions: [
                Padding(
                  padding: EdgeInsets.only(right: isSmallDisplay ? 0 : 4),
                  child: IconButton(
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    splashRadius: isSmallDisplay ? 16 : 23,
                    onPressed: () => selectedTags.isNotEmpty
                        ? showCustomPopup(
                            context,
                            TagsPopup(
                              parentCallback: () {
                                setState(() {
                                  _loadTags = true;
                                });
                              },
                              tagId: selectedTags.first.slug,
                            ),
                            isDismissible: true,
                          )
                        : null,
                    icon: const Icon(
                      Icons.edit,
                    ),
                  ),
                ),
                Padding(
                  padding: EdgeInsets.only(right: isSmallDisplay ? 0 : 8.0),
                  child: IconButton(
                    splashRadius: isSmallDisplay ? 16 : 23,
                    // iconSize: 14,
                    onPressed: () => selectedTags.isNotEmpty
                        ? showCustomPopup(
                            context,
                            DeleteDialog(
                              objName: selectedTags.map((e) {
                                return e.slug;
                              }).toList(),
                              objType: "tags",
                              parentCallback: () {
                                setState(() {
                                  _loadTags = true;
                                });
                              },
                            ),
                            isDismissible: true,
                          )
                        : null,
                    icon: Icon(
                      Icons.delete,
                      color: Colors.red.shade900,
                    ),
                  ),
                ),
                if (isSmallDisplay)
                  IconButton(
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    splashRadius: 16,
                    onPressed: () => showCustomPopup(
                      context,
                      TagsPopup(
                        parentCallback: () {
                          setState(() {
                            _loadTags = true;
                          });
                        },
                      ),
                    ),
                    icon: Icon(
                      Icons.add,
                      color: Colors.blue.shade900,
                    ),
                  )
                else
                  Padding(
                    padding: const EdgeInsets.only(right: 6.0),
                    child: ElevatedButton.icon(
                      onPressed: () => showCustomPopup(
                        context,
                        TagsPopup(
                          parentCallback: () {
                            setState(() {
                              _loadTags = true;
                            });
                          },
                        ),
                      ),
                      icon: const Icon(Icons.add, color: Colors.white),
                      label: Text("${localeMsg.create} Tag"),
                    ),
                  ),
              ],
              rowsPerPage:
                  _tags!.isEmpty ? 1 : (_tags!.length >= 6 ? 6 : _tags!.length),
              columns: [
                DataColumn(
                  label: Text(
                    localeMsg.color,
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
                DataColumn(
                  label: const Text(
                    "Slug",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                  onSort: (columnIndex, ascending) {
                    setState(() {
                      sort = !sort;
                    });
                    onsortColum(columnIndex, ascending);
                  },
                ),
                const DataColumn(
                  label: Text(
                    "Description",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
                const DataColumn(
                  label: Text(
                    "Image",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
              ],
              source: _DataSource(context, _tags!, onTagSelected),
            ),
          ),
        );
      },
    );
  }

  getTags() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchTags();
    switch (result) {
      case Success(value: final value):
        _tags = value;
        _filterTags = _tags;
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        _tags = [];
    }
    _loadTags = false;
  }

  List<Tag> searchTags(String searchText) {
    if (searchText.trim().isEmpty) {
      return _filterTags!.toList();
    }
    switch (_searchField) {
      case TagSearchFields.Description:
        return _filterTags!
            .where((element) => element.description.contains(searchText))
            .toList();
      case TagSearchFields.Slug:
        return _filterTags!
            .where((element) => element.slug.contains(searchText))
            .toList();
      case TagSearchFields.Color:
        return _filterTags!
            .where(
              (element) => element.color
                  .toLowerCase()
                  .contains(searchText.toLowerCase()),
            )
            .toList();
    }
  }

  onTagSelected(int index, bool value) {
    if (index < 0) {
      selectedTags = [];
    } else if (value) {
      selectedTags.add(_tags![index]);
    } else {
      selectedTags.remove(_tags![index]);
    }
  }
}

class _DataSource extends DataTableSource {
  List<Tag> tags;
  final Function onRowSelected;
  _DataSource(this.context, this.tags, this.onRowSelected) {
    _rows = getChildren();
    onRowSelected(-1, false);
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
          notifyListeners();
          onRowSelected(index, value);
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
    for (final tag in tags) {
      final List<DataCell> row = [];
      row.add(colorLabel(tag.color));
      row.add(label(tag.slug, fontWeight: FontWeight.w500));
      row.add(label(tag.description));
      row.add(tag.image.isNotEmpty ? imageLabel(tag.image) : label("-"));
      children.add(CustomRow(row));
    }
    return children;
  }

  DataCell colorLabel(String color) {
    return DataCell(
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: Tooltip(
          message: color,
          child: Icon(Icons.circle, color: Color(int.parse("0xFF$color"))),
        ),
      ),
    );
  }

  DataCell imageLabel(String imagePath) {
    return DataCell(
      SizedBox(
        width: 100,
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Image.network(tenantUrl + imagePath),
        ),
      ),
    );
  }

  DataCell label(String label, {FontWeight fontWeight = FontWeight.w400}) {
    return DataCell(
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 14,
            fontWeight: fontWeight,
          ),
        ),
      ),
    );
  }
}
