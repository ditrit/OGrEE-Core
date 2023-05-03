// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/models/user.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'package:ogree_app/widgets/tenants/tenant_card.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'popups/user_popup.dart';

class UserView extends StatefulWidget {
  Tenant tenant;
  UserView({
    Key? key,
    required this.tenant,
  }) : super(key: key);
  @override
  State<UserView> createState() => _UserViewState();
}

class _UserViewState extends State<UserView> {
  List<User> _users = [];
  late final AppController appController = AppController();
  bool _loadUsers = true;
  List<User> selectedUsers = [];

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return FutureBuilder(
        future: _loadUsers ? getUsers() : null,
        builder: (context, _) {
          if (_users.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          _loadUsers = false;
          return Theme(
            data: ThemeData(
              cardTheme: const CardTheme(
                  elevation: 0,
                  surfaceTintColor: Colors.white,
                  color: Colors.white),
            ),
            child: SingleChildScrollView(
              padding: const EdgeInsets.only(right: 16, top: 0),
              child: PaginatedDataTable(
                checkboxHorizontalMargin: 0,
                header: TextField(
                    decoration: InputDecoration(
                  border: InputBorder.none,
                  isDense: true,
                  label: Text(localeMsg.search),
                  prefixIcon: IconButton(
                    onPressed: () => {},
                    tooltip: "Search",
                    icon: const Icon(
                      Icons.search_rounded,
                    ),
                  ),
                )),
                actions: [
                  Padding(
                    padding: const EdgeInsets.only(right: 4.0),
                    child: IconButton(
                        splashRadius: 23,
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
                                isDismissible: true)
                            : null,
                        icon: const Icon(
                          Icons.edit,
                        )),
                  ),
                  Padding(
                    padding: const EdgeInsets.only(right: 8.0),
                    child: IconButton(
                        splashRadius: 23,
                        // iconSize: 14,
                        onPressed: () => selectedUsers.length > 0
                            ? showCustomPopup(
                                context,
                                DeleteDialog(
                                  objName: selectedUsers.map((e) {
                                    print(e);
                                    return e.id!;
                                  }).toList(),
                                  objType: "users",
                                  parentCallback: () {
                                    setState(() {
                                      _loadUsers = true;
                                    });
                                  },
                                ),
                                isDismissible: true)
                            : null,
                        icon: Icon(
                          Icons.delete,
                          color: Colors.red.shade900,
                        )),
                  ),
                  Padding(
                    padding: const EdgeInsets.only(right: 6.0),
                    child: ElevatedButton.icon(
                      onPressed: () => showCustomPopup(context,
                          UserPopup(parentCallback: () {
                        setState(() {
                          _loadUsers = true;
                        });
                      })),
                      icon: const Icon(Icons.add, color: Colors.white),
                      label: Text("${localeMsg.create} ${localeMsg.user}"),
                    ),
                  ),
                ],
                rowsPerPage: _users.length >= 6 ? 6 : _users.length,
                columns: const [
                  DataColumn(
                      label: Text(
                    "Name",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  )),
                  DataColumn(
                      label: Text(
                    "Email",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  )),
                  DataColumn(
                      label: Text(
                    "Domains (roles)",
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ))
                ],
                source: _DataSource(context, _users, onUserSelected),
              ),
            ),
          );
        });
  }

  getUsers() async {
    _users = await fetchApiUsers(
        "http://${widget.tenant.apiUrl}:${widget.tenant.apiPort}");
  }

  onUserSelected(int index, bool value) {
    if (index < 0) {
      selectedUsers = [];
    } else if (value) {
      selectedUsers.add(_users[index]);
    } else {
      selectedUsers.remove(_users[index]);
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
    List<CustomRow> children = [];
    for (var user in users) {
      List<DataCell> row = [];
      row.add(label(user.name == "null" ? "-" : user.name));
      row.add(label(user.email, fontWeight: FontWeight.w500));
      String domainStr = "";
      for (var domain in user.roles.keys) {
        domainStr =
            "$domainStr ${domain == "*" ? "All domains" : domain} (${user.roles[domain]});";
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
