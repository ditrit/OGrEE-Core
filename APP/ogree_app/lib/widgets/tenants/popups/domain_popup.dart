import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/models/domain.dart';
import 'package:file_picker/file_picker.dart';

class DomainPopup extends StatefulWidget {
  Function() parentCallback;
  String? domainId;
  DomainPopup({super.key, required this.parentCallback, this.domainId});

  @override
  State<DomainPopup> createState() => _DomainPopupState();
}

class _DomainPopupState extends State<DomainPopup>
    with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  String? _domainParent;
  String? _domainName;
  String? _domainColor;
  Color _localColor = Colors.blue.shade900;
  String? _domainDescription;
  bool _isLoading = false;
  bool _isLoadingDelete = false;
  bool _isEdit = false;
  Domain? domain;
  String? domainId;
  late TabController _tabController;
  PlatformFile? _loadedFile;
  String? _loadFileResult;

  @override
  void initState() {
    super.initState();
    _isEdit = widget.domainId != null;
    _tabController = TabController(length: _isEdit ? 1 : 2, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return FutureBuilder(
      future: _isEdit && domain == null ? getDomain() : null,
      builder: (context, _) {
        if (!_isEdit || (_isEdit && domain != null)) {
          return DomainForm(localeMsg);
        } else {
          return const Center(child: CircularProgressIndicator());
        }
      },
    );
  }

  getDomain() async {
    domain = await fetchDomain(widget.domainId!);
    if (domain == null) {
      showSnackBar(context, "Unable to retrieve domain", isError: true);
      Navigator.of(context).pop();
      return;
    }
    domainId = domain!.parent == ""
        ? domain!.name
        : "${domain!.parent}.${domain!.name}";
    _localColor = Color(int.parse("0xFF${domain!.color}"));
  }

  DomainForm(AppLocalizations localeMsg) {
    return Center(
      child: Container(
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.min,
                children: [
                  TabBar(
                    controller: _tabController,
                    // labelPadding: const EdgeInsets.only(left: 20, right: 20),
                    // labelColor: Colors.black,
                    // unselectedLabelColor: Colors.grey,
                    labelStyle: TextStyle(
                        fontSize: 15,
                        fontFamily: GoogleFonts.inter().fontFamily),
                    unselectedLabelStyle: TextStyle(
                        fontSize: 15,
                        fontFamily: GoogleFonts.inter().fontFamily),
                    isScrollable: true,
                    indicatorSize: TabBarIndicatorSize.label,
                    tabs: _isEdit
                        ? [
                            Tab(
                              text: localeMsg.modifyDomain,
                            ),
                          ]
                        : [
                            Tab(
                              text: localeMsg.createDomain,
                            ),
                            Tab(
                              text: localeMsg.createBulkFile,
                            ),
                          ],
                  ),
                  Container(
                    height: 270,
                    child: Padding(
                      padding: const EdgeInsets.only(top: 8.0),
                      child: TabBarView(
                        physics: NeverScrollableScrollPhysics(),
                        controller: _tabController,
                        children: _isEdit
                            ? [
                                getDomainForm(),
                              ]
                            : [
                                getDomainForm(),
                                getBulkFileView(),
                              ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton.icon(
                        style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.blue.shade900),
                        onPressed: () => Navigator.pop(context),
                        label: Text(localeMsg.cancel),
                        icon: const Icon(
                          Icons.cancel_outlined,
                          size: 16,
                        ),
                      ),
                      const SizedBox(width: 15),
                      _isEdit
                          ? Padding(
                              padding: const EdgeInsets.only(right: 15),
                              child: ElevatedButton.icon(
                                style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.red),
                                onPressed: () async {
                                  if (_formKey.currentState!.validate()) {
                                    _formKey.currentState!.save();
                                    setState(() {
                                      _isLoadingDelete = true;
                                    });
                                    var response = await removeObject(
                                        domainId!, "domains");
                                    if (response == "") {
                                      widget.parentCallback();
                                      showSnackBar(context, localeMsg.deleteOK);
                                      Navigator.of(context).pop();
                                    } else {
                                      setState(() {
                                        _isLoadingDelete = false;
                                      });
                                      showSnackBar(context, response,
                                          isError: true);
                                    }
                                  }
                                },
                                label: Text(localeMsg.delete),
                                icon: _isLoadingDelete
                                    ? Container(
                                        width: 24,
                                        height: 24,
                                        padding: const EdgeInsets.all(2.0),
                                        child: const CircularProgressIndicator(
                                          color: Colors.white,
                                          strokeWidth: 3,
                                        ),
                                      )
                                    : const Icon(Icons.delete, size: 16),
                              ),
                            )
                          : Container(),
                      ElevatedButton.icon(
                        onPressed: () async {
                          if (_tabController.index == 1) {
                            if (_loadedFile == null) {
                              showSnackBar(context, localeMsg.mustSelectJSON);
                            } else if (_loadFileResult != null) {
                              widget.parentCallback();
                              Navigator.of(context).pop();
                            } else {
                              var response = await createBulkFile(
                                  _loadedFile!.bytes!, "domains");
                              setState(() {
                                _loadFileResult =
                                    response.replaceAll(",", ",\n");
                                _loadFileResult = _loadFileResult!
                                    .substring(1, _loadFileResult!.length - 1);
                              });
                            }
                          } else {
                            if (_formKey.currentState!.validate()) {
                              _formKey.currentState!.save();
                              setState(() {
                                _isLoading = true;
                              });
                              var newDomain = Domain(
                                  _domainName!,
                                  _domainColor!,
                                  _domainDescription!,
                                  _domainParent!);
                              String response;
                              if (_isEdit) {
                                response =
                                    await updateDomain(domainId!, newDomain);
                              } else {
                                response = await createDomain(newDomain);
                              }
                              if (response == "") {
                                widget.parentCallback();
                                showSnackBar(context,
                                    "${_isEdit ? localeMsg.modifyOK : localeMsg.createOK} 🥳",
                                    isSuccess: true);
                                Navigator.of(context).pop();
                              } else {
                                setState(() {
                                  _isLoading = false;
                                });
                                showSnackBar(context, response, isError: true);
                              }
                            }
                          }
                        },
                        label: Text(_isEdit
                            ? localeMsg.modify
                            : (_loadFileResult == null
                                ? localeMsg.create
                                : "OK")),
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
                  )
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  getDomainForm() {
    final localeMsg = AppLocalizations.of(context)!;
    return ListView(
      children: [
        getFormField(
            save: (newValue) => _domainParent = newValue,
            label: localeMsg.parentDomain,
            icon: Icons.auto_awesome_mosaic,
            initialValue: _isEdit ? domain!.parent : null,
            noValidation: true),
        getFormField(
            save: (newValue) => _domainName = newValue,
            label: localeMsg.domainName,
            icon: Icons.auto_awesome_mosaic,
            initialValue: _isEdit ? domain!.name : null),
        getFormField(
            save: (newValue) => _domainDescription = newValue,
            label: "Description",
            icon: Icons.message,
            initialValue: _isEdit ? domain!.description : null),
        getFormField(
            save: (newValue) => _domainColor = newValue,
            label: localeMsg.color,
            icon: Icons.color_lens,
            formatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9a-fA-F]'))
            ],
            isColor: true,
            initialValue: _isEdit ? domain!.color : null),
      ],
    );
  }

  getBulkFileView() {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: ListView(shrinkWrap: true, children: [
        _loadFileResult == null
            ? Align(
                child: ElevatedButton.icon(
                    onPressed: () async {
                      FilePickerResult? result =
                          await FilePicker.platform.pickFiles();
                      if (result != null) {
                        setState(() {
                          _loadedFile = result.files.single;
                        });
                      }
                    },
                    icon: Icon(Icons.download),
                    label: Text(localeMsg.selectJSON)),
              )
            : Container(),
        _loadedFile != null
            ? Padding(
                padding: const EdgeInsets.only(top: 8.0, bottom: 8.0),
                child: Align(
                  child: Text(localeMsg.fileLoaded(_loadedFile!.name)),
                ),
              )
            : Container(),
        _loadFileResult != null
            ? Container(
                color: Colors.black,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Text(
                    'Result:\n $_loadFileResult',
                    style: TextStyle(color: Colors.white),
                  ),
                ),
              )
            : Container(),
      ]),
    );
  }

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      List<TextInputFormatter>? formatters,
      bool isColor = false,
      String? initialValue,
      bool noValidation = false}) {
    final localeMsg = AppLocalizations.of(context)!;
    return Padding(
      padding: const EdgeInsets.only(left: 2, right: 10),
      child: TextFormField(
        onChanged: isColor
            ? (value) {
                if (value.length == 6) {
                  setState(() {
                    _localColor = Color(int.parse("0xFF$value"));
                  });
                } else {
                  setState(() {
                    _localColor = Colors.blue.shade900;
                  });
                }
              }
            : null,
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (noValidation) {
            return null;
          }
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          if (isColor && text.length < 6) {
            return localeMsg.shouldHaveXChars(6);
          }
          return null;
        },
        maxLength: isColor ? 6 : null,
        inputFormatters: formatters,
        initialValue: initialValue,
        decoration: InputDecoration(
          icon: Icon(icon, color: isColor ? _localColor : Colors.blue.shade900),
          labelText: label,
        ),
      ),
    );
  }
}
