import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
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
  String _imageTag = "latest";

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        // height: 240,
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
                  Row(
                    children: [
                      const Icon(Icons.add_to_photos),
                      Text(
                        localeMsg.newTenant,
                        style: GoogleFonts.inter(
                          fontSize: 22,
                          color: Colors.black,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 45),
                  getFormField(
                      save: (newValue) => _tenantName = newValue,
                      label: localeMsg.tenantName,
                      icon: Icons.business_center),
                  getFormField(
                      save: (newValue) => _tenantPassword = newValue,
                      label: localeMsg.tenantPassword,
                      icon: Icons.lock),
                  const SizedBox(height: 8),
                  Wrap(
                    alignment: WrapAlignment.start,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Padding(
                        padding: const EdgeInsets.only(right: 12),
                        child: Text("Services:"),
                      ),
                      getCheckBox("API", true, (_) {}, enabled: false),
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
                  getFormField(
                      save: (newValue) => _imageTag = newValue!,
                      label: "Version du déploiement (tag)",
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
                          label: "${localeMsg.webUrl} (hostname:port)",
                          icon: Icons.monitor,
                          prefix: "http://",
                          isUrl: true,
                        )
                      : Container(),
                  _hasWeb
                      ? Padding(
                          padding: const EdgeInsets.only(top: 8.0),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              Padding(
                                padding: const EdgeInsets.only(right: 20),
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
                                        await FilePicker.platform.pickFiles(
                                            type: FileType.custom,
                                            allowedExtensions: ["png"],
                                            withData: true);
                                    if (result != null) {
                                      setState(() {
                                        _loadedImage = result.files.single;
                                      });
                                    }
                                  },
                                  icon: const Icon(Icons.download),
                                  label: Text(localeMsg.selectLogo)),
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
                          label: "${localeMsg.docUrl} (hostname:port)",
                          icon: Icons.book,
                          prefix: "http://",
                          isUrl: true,
                        )
                      : Container(),
                  const SizedBox(height: 30),
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
                                showSnackBar(
                                    context, "${localeMsg.tenantCreated} 🥳",
                                    isSuccess: true);
                                Navigator.of(context).pop();
                              } else {
                                setState(() {
                                  _isLoading = false;
                                });
                                showSnackBar(context, response, isError: true);
                              }
                            }
                          },
                          label: Text(localeMsg.create),
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
                              : const Icon(Icons.check_circle, size: 16))
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

  getCheckBox(String title, bool value, Function(bool?) onChange,
      {bool enabled = true}) {
    return SizedBox(
      width: 100,
      child: CheckboxListTile(
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
      padding: const EdgeInsets.only(left: 2, right: 10),
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
        decoration: InputDecoration(
          icon: Icon(icon, color: Colors.blue.shade900),
          labelText: label,
          prefixText: prefix,
          suffixText: suffix,
        ),
      ),
    );
  }
}
