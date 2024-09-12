// ignore_for_file: constant_identifier_names

import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/user.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/common/delete_dialog_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/user_popup.dart';

enum UserSearchFields { Name, Email, Domain, Role }

class UserView extends StatefulWidget {
  UserSearchFields searchField;
  String? searchText;
  Function? parentCallback;
  UserView({
    super.key,
    this.searchField = UserSearchFields.Name,
    this.searchText,
    this.parentCallback,
  });
  @override
  State<UserView> createState() => _UserViewState();
}

class _UserViewState extends State<UserView> {
  List<User>? _users;
  bool _loadUsers = true;
  List<User> selectedUsers = [];
  List<User>? _filterUsers;
  bool sort = true;
  UserSearchFields _searchField = UserSearchFields.Name;

  @override
  void initState() {
    super.initState();
    _searchField = widget.searchField;
    _loadUsers = true;
  }

  onsortColum(int columnIndex, bool ascending) {
    if (columnIndex == 1) {
      if (ascending) {
        _users!.sort((a, b) => a.email.compareTo(b.email));
      } else {
        _users!.sort((a, b) => b.email.compareTo(a.email));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return FutureBuilder(
      future: _loadUsers ? getUsers() : null,
      builder: (context, _) {
        if (_users == null) {
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
                    child: DropdownButtonFormField<UserSearchFields>(
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
                      items: UserSearchFields.values
                          .map<DropdownMenuItem<UserSearchFields>>(
                              (UserSearchFields value) {
                        return DropdownMenuItem<UserSearchFields>(
                          value: value,
                          child: Text(
                            value.name,
                            overflow: TextOverflow.ellipsis,
                          ),
                        );
                      }).toList(),
                      onChanged: (UserSearchFields? value) {
                        setState(() {
                          _searchField = value!;
                        });
                      },
                    ),
                  ),
                  const SizedBox(width: 8),
                  SizedBox(
                    width: 225,
                    child: TextFormField(
                      textAlignVertical: TextAlignVertical.center,
                      initialValue: widget.searchText,
                      onChanged: (value) {
                        setState(() {
                          _users = searchUsers(value);
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
                    onPressed: () => selectedUsers.isNotEmpty
                        ? showCustomPopup(
                            context,
                            UserPopup(
                              parentCallback: () {
                                setState(() {
                                  _loadUsers = true;
                                });
                              },
                              modifyUser: selectedUsers.first,
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
                    onPressed: () => selectedUsers.isNotEmpty
                        ? showCustomPopup(
                            context,
                            DeleteDialog(
                              objName: selectedUsers.map((e) {
                                return e.id!;
                              }).toList(),
                              objType: "users",
                              parentCallback: () {
                                setState(() {
                                  _loadUsers = true;
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
                      UserPopup(
                        parentCallback: () {
                          setState(() {
                            _loadUsers = true;
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
                        UserPopup(
                          parentCallback: () {
                            setState(() {
                              _loadUsers = true;
                            });
                          },
                        ),
                      ),
                      icon: const Icon(Icons.add, color: Colors.white),
                      label: Text("${localeMsg.create} ${localeMsg.user}"),
                    ),
                  ),
              ],
              rowsPerPage: _users!.isEmpty
                  ? 1
                  : (_users!.length >= 6 ? 6 : _users!.length),
              columns: [
                const DataColumn(
                  label: Text(
                    "Name",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
                DataColumn(
                  label: const Text(
                    "Email",
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
                    "Domains (roles)",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
              ],
              source: _DataSource(context, _users!, onUserSelected),
            ),
          ),
        );
      },
    );
  }

  getUsers() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchApiUsers();
    switch (result) {
      case Success(value: final value):
        _users = value;
        _filterUsers = _users;
        if (widget.searchText != null && widget.searchText!.isNotEmpty) {
          // search filter set by parent widget
          _users = searchUsers(widget.searchText!);
          // let him know it has been applied
          widget.parentCallback!();
        }
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        _users = [];
    }
    _loadUsers = false;
  }

  List<User> searchUsers(String searchText) {
    if (searchText.trim().isEmpty) {
      return _filterUsers!.toList();
    }
    switch (_searchField) {
      case UserSearchFields.Name:
        return _filterUsers!
            .where((element) => element.name.contains(searchText))
            .toList();
      case UserSearchFields.Email:
        return _filterUsers!
            .where((element) => element.email.contains(searchText))
            .toList();
      case UserSearchFields.Domain:
        return _filterUsers!.where((element) {
          for (final domain in element.roles.keys) {
            if (domain.contains(searchText) || domain == allDomainsConvert) {
              return true;
            }
          }
          return false;
        }).toList();
      case UserSearchFields.Role:
        return _filterUsers!.where((element) {
          for (final role in element.roles.values) {
            if (role.contains(searchText)) {
              return true;
            }
          }
          return false;
        }).toList();
    }
  }

  onUserSelected(int index, bool value) {
    if (index < 0) {
      selectedUsers = [];
    } else if (value) {
      selectedUsers.add(_users![index]);
    } else {
      selectedUsers.remove(_users![index]);
    }
  }
}

class _DataSource extends DataTableSource {
  List<User> users;
  final Function onRowSelected;
  _DataSource(this.context, this.users, this.onRowSelected) {
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
    for (final user in users) {
      final List<DataCell> row = [];
      row.add(label(user.name == "null" ? "-" : user.name));
      row.add(label(user.email, fontWeight: FontWeight.w500));
      String domainStr = "";
      for (final domain in user.roles.keys) {
        domainStr = "$domainStr $domain (${user.roles[domain]});";
      }
      row.add(label(domainStr));
      children.add(CustomRow(row));
    }
    return children;
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
