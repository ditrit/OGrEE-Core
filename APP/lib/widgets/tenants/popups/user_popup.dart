import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/user.dart';
import 'package:ogree_app/widgets/common/form_field.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';

class UserPopup extends StatefulWidget {
  final Function() parentCallback;
  final User? modifyUser;
  const UserPopup({super.key, required this.parentCallback, this.modifyUser});

  @override
  State<UserPopup> createState() => _UserPopupState();
}

class _UserPopupState extends State<UserPopup> with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  String? _userName;
  String? _userEmail;
  String? _userPassword;
  bool _isLoading = false;
  List<String>? domainList;
  List<String> selectedDomain = [];
  List<String> roleList = <String>['Manager', 'User', 'Viewer'];
  List<String> selectedRole = [];
  List<Widget> domainRoleRows = [];
  bool _isEdit = false;
  late TabController _tabController;
  PlatformFile? _loadedFile;
  String? _loadFileResult;
  bool _isSmallDisplay = false;

  @override
  void initState() {
    super.initState();
    _isEdit = widget.modifyUser != null;
    _tabController = TabController(length: _isEdit ? 1 : 2, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return FutureBuilder(
      future: domainList == null ? getDomains() : null,
      builder: (context, _) {
        if (domainList == null) {
          return const Center(child: CircularProgressIndicator());
        }
        return Center(
          child: Container(
            width: 500,
            margin: const EdgeInsets.symmetric(horizontal: 10),
            decoration: PopupDecoration,
            child: Padding(
              padding: EdgeInsets.fromLTRB(
                _isSmallDisplay ? 20 : 40,
                8,
                _isSmallDisplay ? 20 : 40,
                15,
              ),
              child: Material(
                color: Colors.white,
                child: Form(
                  key: _formKey,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      TabBar(
                        tabAlignment: TabAlignment.center,
                        controller: _tabController,
                        labelStyle: TextStyle(
                          fontSize: 15,
                          fontFamily: GoogleFonts.inter().fontFamily,
                        ),
                        unselectedLabelStyle: TextStyle(
                          fontSize: 15,
                          fontFamily: GoogleFonts.inter().fontFamily,
                        ),
                        isScrollable: true,
                        indicatorSize: TabBarIndicatorSize.label,
                        tabs: _isEdit
                            ? [
                                Tab(
                                  text: localeMsg.modifyUser,
                                ),
                              ]
                            : [
                                Tab(
                                  text: localeMsg.createUser,
                                ),
                                Tab(
                                  text: localeMsg.createBulkFile,
                                ),
                              ],
                      ),
                      SizedBox(
                        height: 320,
                        child: Padding(
                          padding: const EdgeInsets.only(top: 16.0),
                          child: TabBarView(
                            physics: const NeverScrollableScrollPhysics(),
                            controller: _tabController,
                            children: _isEdit
                                ? [
                                    getUserView(localeMsg),
                                  ]
                                : [
                                    getUserView(localeMsg),
                                    getBulkFileView(localeMsg),
                                  ],
                          ),
                        ),
                      ),
                      const SizedBox(height: 10),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton.icon(
                            style: OutlinedButton.styleFrom(
                              foregroundColor: Colors.blue.shade900,
                            ),
                            onPressed: () => Navigator.pop(context),
                            label: Text(localeMsg.cancel),
                            icon: const Icon(
                              Icons.cancel_outlined,
                              size: 16,
                            ),
                          ),
                          const SizedBox(width: 15),
                          ElevatedButton.icon(
                            onPressed: () => userAction(localeMsg),
                            label: Text(
                              _isEdit
                                  ? localeMsg.modify
                                  : (_loadFileResult == null
                                      ? localeMsg.create
                                      : "OK"),
                            ),
                            icon: _isLoading
                                ? Container(
                                    width: 24,
                                    height: 24,
                                    padding: const EdgeInsets.all(2.0),
                                    child: const CircularProgressIndicator(
                                      color: Colors.white,
                                      strokeWidth: 3,
                                    ),
                                  )
                                : const Icon(Icons.check_circle, size: 16),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Future<void> getDomains() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchObjectsTree(
      namespace: Namespace.Organisational,
      isTenantMode: true,
    );
    switch (result) {
      case Success(value: final listValue):
        domainList = listValue[0]
            .values
            .reduce((value, element) => List.from(value + element));
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        if (mounted) Navigator.pop(context);
        return;
    }

    if (!_isEdit) {
      if (domainList!.isNotEmpty) {
        domainList!.add(allDomainsConvert);
        domainRoleRows.add(addDomainRoleRow(0));
      }
    } else {
      domainList!.add(allDomainsConvert);
      final roles = widget.modifyUser!.roles;
      for (var i = 0; i < roles.length; i++) {
        selectedDomain.add(roles.keys.elementAt(i));
        selectedRole.add(roles.values.elementAt(i).capitalize());
        domainRoleRows.add(addDomainRoleRow(i, useDefaultValue: false));
      }
    }
  }

  Map<String, String> getRolesMap() {
    final Map<String, String> roles = {};
    for (var i = 0; i < selectedDomain.length; i++) {
      if (roles.containsKey(selectedDomain[i])) {
        throw Exception(AppLocalizations.of(context)!.onlyOneRoleDomain);
      }
      roles[selectedDomain[i]] = selectedRole[i].toLowerCase();
    }
    return roles;
  }

  ListView getUserView(AppLocalizations localeMsg) {
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        CustomFormField(
          save: (newValue) => _userName = newValue,
          label: "Name",
          icon: Icons.person,
          initialValue: _isEdit ? widget.modifyUser!.name : null,
        ),
        CustomFormField(
          save: (newValue) => _userEmail = newValue,
          label: "Email",
          icon: Icons.alternate_email,
          initialValue: _isEdit ? widget.modifyUser!.email : null,
          isReadOnly: _isEdit,
        ),
        CustomFormField(
          save: (newValue) => _userPassword = newValue,
          label: localeMsg.password,
          icon: Icons.lock,
          initialValue: _isEdit ? widget.modifyUser!.password : null,
          isObscure: true,
          isReadOnly: _isEdit,
        ),
        const Padding(
          padding: EdgeInsets.only(top: 8.0, bottom: 10, left: 12),
          child: Text("Permissions"),
        ),
        Padding(
          padding: const EdgeInsets.only(left: 4),
          child: Column(children: domainRoleRows),
        ),
        Padding(
          padding: const EdgeInsets.only(left: 6),
          child: Align(
            alignment: Alignment.bottomLeft,
            child: TextButton.icon(
              onPressed: () => setState(() {
                domainRoleRows.add(addDomainRoleRow(domainRoleRows.length));
              }),
              icon: const Icon(Icons.add),
              label: Text(localeMsg.domain),
            ),
          ),
        ),
      ],
    );
  }

  Center getBulkFileView(AppLocalizations localeMsg) {
    return Center(
      child: ListView(
        shrinkWrap: true,
        children: [
          if (_loadFileResult == null)
            Align(
              child: ElevatedButton.icon(
                onPressed: () async {
                  final FilePickerResult? result =
                      await FilePicker.platform.pickFiles(withData: true);
                  if (result != null) {
                    setState(() {
                      _loadedFile = result.files.single;
                    });
                  }
                },
                icon: const Icon(Icons.download),
                label: Text(localeMsg.selectJSON),
              ),
            )
          else
            Container(),
          if (_loadedFile != null)
            Padding(
              padding: const EdgeInsets.only(top: 8.0, bottom: 8.0),
              child: Align(
                child: Text(localeMsg.fileLoaded(_loadedFile!.name)),
              ),
            )
          else
            Container(),
          if (_loadFileResult != null)
            Container(
              color: Colors.black,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: Text(
                  'Result:\n$_loadFileResult',
                  style: const TextStyle(color: Colors.white),
                ),
              ),
            )
          else
            Container(),
        ],
      ),
    );
  }

  rebuildDomainRole() {
    domainRoleRows = [];
    for (var i = 0; i < selectedDomain.length; i++) {
      domainRoleRows.add(addDomainRoleRow(i, useDefaultValue: false));
    }
  }

  removeDomainRoleRow(int rowIdx) {
    selectedDomain.removeAt(rowIdx);
    selectedRole.removeAt(rowIdx);
    rebuildDomainRole();
  }

  StatefulBuilder addDomainRoleRow(int rowIdx, {bool useDefaultValue = true}) {
    if (useDefaultValue) {
      selectedDomain.add(domainList!.first);
      selectedRole.add(roleList.first);
    }
    return StatefulBuilder(
      builder: (context, localSetState) {
        return Padding(
          padding: const EdgeInsets.only(top: 4.0),
          child: Row(
            children: [
              Flexible(
                flex: 3,
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    color: const Color.fromARGB(255, 248, 247, 247),
                    borderRadius: BorderRadius.circular(12.0),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 12.0),
                    child: DropdownButton<String>(
                      borderRadius: BorderRadius.circular(12.0),
                      underline: Container(),
                      style: const TextStyle(fontSize: 14, color: Colors.black),
                      isExpanded: true,
                      value: selectedDomain[rowIdx],
                      items: domainList!
                          .map<DropdownMenuItem<String>>((String value) {
                        return DropdownMenuItem<String>(
                          value: value,
                          child: Text(value),
                        );
                      }).toList(),
                      onChanged: (String? value) {
                        localSetState(() {
                          selectedDomain[rowIdx] = value!;
                        });
                      },
                    ),
                  ),
                ),
              ),
              Padding(
                padding: EdgeInsets.symmetric(
                    horizontal: _isSmallDisplay ? 0 : 16.0),
                child: Icon(
                  Icons.arrow_forward,
                  color: Colors.blue.shade600,
                ),
              ),
              Flexible(
                flex: 2,
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    color: const Color.fromARGB(255, 248, 247, 247),
                    borderRadius: BorderRadius.circular(12.0),
                  ),
                  child: Padding(
                    padding: EdgeInsets.symmetric(
                      horizontal: _isSmallDisplay ? 6 : 12.0,
                    ),
                    child: DropdownButton<String>(
                      borderRadius: BorderRadius.circular(12.0),
                      underline: Container(),
                      style: const TextStyle(fontSize: 14, color: Colors.black),
                      isExpanded: true,
                      value: selectedRole[rowIdx],
                      items: roleList
                          .map<DropdownMenuItem<String>>((String value) {
                        return DropdownMenuItem<String>(
                          value: value,
                          child: Text(value),
                        );
                      }).toList(),
                      onChanged: (String? value) {
                        localSetState(() {
                          selectedRole[rowIdx] = value!;
                        });
                      },
                    ),
                  ),
                ),
              ),
              if (rowIdx > 0)
                IconButton(
                  padding: const EdgeInsets.all(4),
                  constraints: const BoxConstraints(),
                  iconSize: 14,
                  onPressed: () => setState(() => removeDomainRoleRow(rowIdx)),
                  icon: Icon(
                    Icons.delete,
                    color: Colors.red.shade400,
                  ),
                )
              else
                const SizedBox(width: 22),
            ],
          ),
        );
      },
    );
  }

  userAction(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (_tabController.index == 1) {
      if (_loadedFile == null) {
        showSnackBar(
          messenger,
          localeMsg.mustSelectJSON,
        );
      } else if (_loadFileResult != null) {
        widget.parentCallback();
        Navigator.of(context).pop();
      } else {
        final result = await createBulkFile(
          _loadedFile!.bytes!,
          "users",
        );
        switch (result) {
          case Success(value: final value):
            setState(() {
              _loadFileResult =
                  value.replaceAll("},", "},\n> ").replaceFirst("{", ">  ");
            });
          case Failure(exception: final exception):
            showSnackBar(
              messenger,
              exception.toString(),
              isError: true,
            );
        }
      }
    } else {
      if (_formKey.currentState!.validate()) {
        _formKey.currentState!.save();
        try {
          final Map<String, String> roles = getRolesMap();
          setState(() {
            _isLoading = true;
          });

          Result response;
          if (_isEdit) {
            response = await modifyUser(
              widget.modifyUser!.id!,
              roles,
            );
          } else {
            response = await createUser(
              User(
                name: _userName!,
                email: _userEmail!,
                password: _userPassword!,
                roles: roles,
              ),
            );
          }

          switch (response) {
            case Success():
              widget.parentCallback();
              showSnackBar(
                messenger,
                _isEdit ? localeMsg.modifyOK : localeMsg.createOK,
                isSuccess: true,
              );
              if (mounted) {
                Navigator.of(context).pop();
              }
            case Failure(exception: final exception):
              setState(() {
                _isLoading = false;
              });
              showSnackBar(
                messenger,
                exception.toString(),
                isError: true,
              );
          }
        } catch (e) {
          showSnackBar(
            messenger,
            e.toString(),
            isError: true,
          );
          return;
        }
      }
    }
  }
}
