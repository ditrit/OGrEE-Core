import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/domain.dart';
import 'package:ogree_app/widgets/common/form_field.dart';

class DomainPopup extends StatefulWidget {
  Function() parentCallback;
  String? domainId;
  String? parentId;
  DomainPopup({
    super.key,
    required this.parentCallback,
    this.domainId,
    this.parentId,
  });

  @override
  State<DomainPopup> createState() => _DomainPopupState();
}

class _DomainPopupState extends State<DomainPopup>
    with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  String? _domainParent;
  String? _domainName;
  String? _domainColor;
  String? _domainDescription;
  bool _isLoading = false;
  bool _isLoadingDelete = false;
  bool _isEdit = false;
  Domain? domain;
  String? domainId;
  late TabController _tabController;
  PlatformFile? _loadedFile;
  String? _loadFileResult;
  bool _isSmallDisplay = false;
  final colorTextController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _isEdit = widget.domainId != null;
    _tabController = TabController(length: _isEdit ? 1 : 2, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

    return FutureBuilder(
      future: _isEdit && domain == null ? getDomain(localeMsg) : null,
      builder: (context, _) {
        if (!_isEdit || (_isEdit && domain != null)) {
          return domainForm(localeMsg);
        } else {
          return const Center(child: CircularProgressIndicator());
        }
      },
    );
  }

  Future<void> getDomain(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchDomain(widget.domainId!);
    switch (result) {
      case Success(value: final value):
        domain = value;
        domainId = domain!.parent == ""
            ? domain!.name
            : "${domain!.parent}.${domain!.name}";
        colorTextController.text = domain!.color;
      case Failure():
        showSnackBar(messenger, localeMsg.noDomain, isError: true);
        if (mounted) Navigator.of(context).pop();
        return;
    }
  }

  Center domainForm(AppLocalizations localeMsg) {
    return Center(
      child: Container(
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
            _isSmallDisplay ? 30 : 40,
            8,
            _isSmallDisplay ? 30 : 40,
            15,
          ),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
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
                  SizedBox(
                    height: 270,
                    child: Padding(
                      padding: const EdgeInsets.only(top: 16.0),
                      child: TabBarView(
                        physics: const NeverScrollableScrollPhysics(),
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
                  const SizedBox(height: 5),
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
                      if (_isEdit)
                        TextButton.icon(
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.red.shade900,
                          ),
                          onPressed: () => removeDomain(localeMsg),
                          label: Text(_isSmallDisplay ? "" : localeMsg.delete),
                          icon: _isLoadingDelete
                              ? Container(
                                  width: 24,
                                  height: 24,
                                  padding: const EdgeInsets.all(2.0),
                                  child: CircularProgressIndicator(
                                    color: Colors.red.shade900,
                                    strokeWidth: 3,
                                  ),
                                )
                              : const Icon(
                                  Icons.delete,
                                  size: 16,
                                ),
                        )
                      else
                        Container(),
                      if (_isSmallDisplay)
                        Container()
                      else
                        const SizedBox(width: 10),
                      ElevatedButton.icon(
                        onPressed: () async {
                          final messenger = ScaffoldMessenger.of(context);
                          if (_tabController.index == 1) {
                            if (_loadedFile == null) {
                              showSnackBar(
                                ScaffoldMessenger.of(context),
                                localeMsg.mustSelectJSON,
                              );
                            } else if (_loadFileResult != null) {
                              widget.parentCallback();
                              Navigator.of(context).pop();
                            } else {
                              final result = await createBulkFile(
                                _loadedFile!.bytes!,
                                "domains",
                              );
                              switch (result) {
                                case Success(value: final value):
                                  setState(() {
                                    _loadFileResult =
                                        value.replaceAll(",", ",\n");
                                    _loadFileResult =
                                        _loadFileResult!.substring(
                                      1,
                                      _loadFileResult!.length - 1,
                                    );
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
                              setState(() {
                                _isLoading = true;
                              });
                              final newDomain = Domain(
                                _domainName!,
                                _domainColor!,
                                _domainDescription!,
                                _domainParent!,
                              );
                              Result result;
                              if (_isEdit) {
                                result =
                                    await updateDomain(domainId!, newDomain);
                              } else {
                                result = await createDomain(newDomain);
                              }
                              switch (result) {
                                case Success():
                                  widget.parentCallback();
                                  showSnackBar(
                                    messenger,
                                    "${_isEdit ? localeMsg.modifyOK : localeMsg.createOK} 🥳",
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
                            }
                          }
                        },
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
  }

  removeDomain(AppLocalizations localeMsg) async {
    setState(() {
      _isLoadingDelete = true;
    });
    final messenger = ScaffoldMessenger.of(context);
    final result = await removeObject(domainId!, "domains");
    switch (result) {
      case Success():
        widget.parentCallback();
        showSnackBar(messenger, localeMsg.deleteOK);
        if (mounted) {
          Navigator.of(context).pop();
        }
      case Failure(exception: final exception):
        setState(() {
          _isLoadingDelete = false;
        });
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }

  ListView getDomainForm() {
    final localeMsg = AppLocalizations.of(context)!;
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        CustomFormField(
          save: (newValue) => _domainParent = newValue,
          label: localeMsg.parentDomain,
          icon: Icons.auto_awesome_mosaic,
          initialValue: _isEdit ? domain!.parent : widget.parentId,
          shouldValidate: false,
        ),
        CustomFormField(
          save: (newValue) => _domainName = newValue,
          label: localeMsg.domainName,
          icon: Icons.auto_awesome_mosaic,
          initialValue: _isEdit ? domain!.name : null,
        ),
        CustomFormField(
          save: (newValue) => _domainDescription = newValue,
          label: "Description",
          icon: Icons.message,
          initialValue: _isEdit ? domain!.description : null,
        ),
        CustomFormField(
          save: (newValue) => _domainColor = newValue,
          label: localeMsg.color,
          icon: Icons.color_lens,
          isColor: true,
          colorTextController: colorTextController,
          maxLength: 6,
        ),
      ],
    );
  }

  Center getBulkFileView() {
    final localeMsg = AppLocalizations.of(context)!;
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
                  'Result:\n $_loadFileResult',
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
}
