import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tenant.dart';

class CreateTenantPopup extends StatefulWidget {
  Function() parentCallback;
  CreateTenantPopup({super.key, required this.parentCallback});

  @override
  State<CreateTenantPopup> createState() => _CreateTenantPopupState();
}

class _CreateTenantPopupState extends State<CreateTenantPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _tenantName;
  String? _tenantPassword;
  String? _apiUrl;
  String _webUrl = "";
  String? _apiPort;
  String _webPort = "";
  String _docUrl = "";
  String _docPort = "";
  bool _hasWeb = true;
  bool _hasDoc = false;
  bool _isLoading = false;
  PlatformFile? _loadedImage;
  String _imageTag = "main";
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 540),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Form(
            key: _formKey,
            child: ScaffoldMessenger(
                child: Builder(
                    builder: (context) => Scaffold(
                          backgroundColor: Colors.white,
                          body: ListView(
                            shrinkWrap: true,
                            children: [
                              Center(
                                  child: Text(
                                localeMsg.newTenant,
                                style:
                                    Theme.of(context).textTheme.headlineMedium,
                              )),
                              // const Divider(height: 35),
                              const SizedBox(height: 20),
                              getFormField(
                                  save: (newValue) => _tenantName = newValue,
                                  label: localeMsg.tenantName,
                                  icon: Icons.business_center),
                              getFormField(
                                  save: (newValue) =>
                                      _tenantPassword = newValue,
                                  label: localeMsg.tenantPassword,
                                  icon: Icons.lock),
                              const SizedBox(height: 4),
                              Wrap(
                                alignment: WrapAlignment.center,
                                crossAxisAlignment: WrapCrossAlignment.center,
                                children: [
                                  getCheckBox("API", true, (_) {},
                                      enabled: false),
                                  getCheckBox(
                                      "WEB",
                                      _hasWeb,
                                      (value) => setState(() {
                                            _hasWeb = value!;
                                          })),
                                  getCheckBox(
                                      "DOC",
                                      _hasDoc,
                                      (value) => setState(() {
                                            _hasDoc = value!;
                                          })),
                                ],
                              ),
                              const SizedBox(height: 10),
                              getFormField(
                                  save: (newValue) => _imageTag = newValue!,
                                  label: "Version du déploiement (branch)",
                                  icon: Icons.access_time,
                                  initial: _imageTag),
                              getFormField(
                                save: (newValue) {
                                  var splitted = newValue!.split(":");
                                  _apiUrl = splitted[0];
                                  _apiPort = splitted[1];
                                },
                                label: "${localeMsg.apiUrl} (hostname:port)",
                                icon: Icons.cloud,
                                prefix: "http://",
                                isUrl: true,
                              ),
                              _hasWeb
                                  ? getFormField(
                                      save: (newValue) {
                                        var splitted = newValue!.split(":");
                                        _webUrl = splitted[0];
                                        _webPort = splitted[1];
                                      },
                                      label:
                                          "${localeMsg.webUrl} (hostname:port)",
                                      icon: Icons.monitor,
                                      prefix: "http://",
                                      isUrl: true,
                                    )
                                  : Container(),
                              _hasWeb
                                  ? Padding(
                                      padding: const EdgeInsets.only(
                                          top: 8.0, bottom: 8),
                                      child: Wrap(
                                        alignment: WrapAlignment.end,
                                        crossAxisAlignment:
                                            WrapCrossAlignment.center,
                                        children: [
                                          Padding(
                                            padding: const EdgeInsets.only(
                                                right: 20),
                                            child: _loadedImage == null
                                                ? Image.asset(
                                                    "assets/custom/logo.png",
                                                    height: 40,
                                                  )
                                                : Image.memory(
                                                    _loadedImage!.bytes!,
                                                    height: 40,
                                                  ),
                                          ),
                                          ElevatedButton.icon(
                                              onPressed: () async {
                                                FilePickerResult? result =
                                                    await FilePicker.platform
                                                        .pickFiles(
                                                            type:
                                                                FileType.custom,
                                                            allowedExtensions: [
                                                              "png"
                                                            ],
                                                            withData: true);
                                                if (result != null) {
                                                  setState(() {
                                                    _loadedImage =
                                                        result.files.single;
                                                  });
                                                }
                                              },
                                              icon: const Icon(Icons.download),
                                              label:
                                                  Text(localeMsg.selectLogo)),
                                        ],
                                      ),
                                    )
                                  : Container(),
                              _hasDoc
                                  ? getFormField(
                                      save: (newValue) {
                                        var splitted = newValue!.split(":");
                                        _docUrl = splitted[0];
                                        _docPort = splitted[1];
                                      },
                                      label:
                                          "${localeMsg.docUrl} (hostname:port)",
                                      icon: Icons.book,
                                      prefix: "http://",
                                      isUrl: true,
                                    )
                                  : Container(),
                              const SizedBox(height: 10),
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
                                  ElevatedButton.icon(
                                      onPressed: () async {
                                        if (_formKey.currentState!.validate()) {
                                          _formKey.currentState!.save();
                                          setState(() {
                                            _isLoading = true;
                                          });
                                          // Load logo first, if provided
                                          String response = localeMsg.notLoaded;
                                          if (_loadedImage != null) {
                                            response = await uploadImage(
                                                _loadedImage!, _tenantName!);
                                            print(response);
                                            if (response != "") {
                                              showSnackBar(context,
                                                  "${localeMsg.failedToUpload} $response");
                                            }
                                          }
                                          // Create tenant
                                          response = await createTenant(Tenant(
                                              _tenantName!,
                                              _tenantPassword!,
                                              _apiUrl!,
                                              _webUrl,
                                              _apiPort!,
                                              _webPort,
                                              _hasWeb,
                                              _hasDoc,
                                              _docUrl,
                                              _docPort,
                                              _imageTag));
                                          if (response == "") {
                                            widget.parentCallback();
                                            showSnackBar(context,
                                                "${localeMsg.tenantCreated} 🥳",
                                                isSuccess: true);
                                            Navigator.of(context).pop();
                                          } else {
                                            setState(() {
                                              _isLoading = false;
                                            });
                                            showSnackBar(context, response,
                                                isError: true);
                                          }
                                        }
                                      },
                                      label: Text(localeMsg.create),
                                      icon: _isLoading
                                          ? Container(
                                              width: 24,
                                              height: 24,
                                              padding:
                                                  const EdgeInsets.all(2.0),
                                              child:
                                                  const CircularProgressIndicator(
                                                color: Colors.white,
                                                strokeWidth: 3,
                                              ),
                                            )
                                          : const Icon(Icons.check_circle,
                                              size: 16))
                                ],
                              )
                            ],
                          ),
                        ))),
          ),
        ),
      ),
    );
  }

  getCheckBox(String title, bool value, Function(bool?) onChange,
      {bool enabled = true}) {
    return SizedBox(
      width: 100,
      child: CheckboxListTile(
        activeColor: Colors.blue.shade600,
        contentPadding: EdgeInsets.zero,
        controlAffinity: ListTileControlAffinity.leading,
        value: value,
        enabled: enabled,
        onChanged: (value) => onChange(value),
        title: Transform.translate(
            offset: const Offset(-10, 0), child: Text(title)),
      ),
    );
  }

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      String? initial,
      bool isUrl = false}) {
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
        initialValue: initial,
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          if (isUrl) {
            var splitted = text.split(":");
            if (splitted.length != 2) {
              return AppLocalizations.of(context)!.wrongFormatUrl;
            }
            if (int.tryParse(splitted[1]) == null) {
              return AppLocalizations.of(context)!.wrongFormatPort;
            }
          }
          return null;
        },
        inputFormatters: formatters,
        decoration: GetFormInputDecoration(_isSmallDisplay, label,
            prefixText: prefix, suffixText: suffix, icon: icon),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
