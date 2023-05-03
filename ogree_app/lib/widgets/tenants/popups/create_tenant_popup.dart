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
  String? _webUrl;
  String? _apiPort;
  String? _webPort;
  bool _hasWeb = true;
  bool _hasCli = true;
  bool _isLoading = false;

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
                      label: "Tenant Name",
                      icon: Icons.business_center),
                  getFormField(
                      save: (newValue) => _tenantPassword = newValue,
                      label: "Tenant Admin Password",
                      icon: Icons.lock),
                  const SizedBox(height: 8),
                  Wrap(
                    alignment: WrapAlignment.start,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Text("Services:"),
                      SizedBox(
                        width: 112,
                        child: CheckboxListTile(
                          controlAffinity: ListTileControlAffinity.leading,
                          value: true,
                          enabled: false,
                          onChanged: (_) {},
                          title: Text("API"),
                        ),
                      ),
                      SizedBox(
                        width: 120,
                        child: CheckboxListTile(
                          controlAffinity: ListTileControlAffinity.leading,
                          value: _hasWeb,
                          onChanged: (value) => setState(() {
                            _hasWeb = value!;
                          }),
                          title: Text("WEB"),
                        ),
                      ),
                      SizedBox(
                        width: 112,
                        child: CheckboxListTile(
                          controlAffinity: ListTileControlAffinity.leading,
                          value: _hasCli,
                          onChanged: (value) => setState(() {
                            _hasCli = value!;
                          }),
                          title: Text("CLI"),
                        ),
                      ),
                    ],
                  ),
                  getFormField(
                    save: (newValue) {
                      var splitted = newValue!.split(":");
                      _apiUrl = splitted[0];
                      _apiPort = splitted[1];
                    },
                    label: "New API URL (hostname:port)",
                    icon: Icons.cloud,
                    prefix: "http://",
                    isUrl: true,
                  ),
                  getFormField(
                    save: (newValue) {
                      var splitted = newValue!.split(":");
                      _webUrl = splitted[0];
                      _webPort = splitted[1];
                    },
                    label: "New Web URL (hostname:port)",
                    icon: Icons.monitor,
                    prefix: "http://",
                    isUrl: true,
                  ),
                  const SizedBox(height: 40),
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
                              var response = await createTenant(Tenant(
                                  _tenantName!,
                                  _tenantPassword!,
                                  _apiUrl!,
                                  _webUrl!,
                                  _apiPort!,
                                  _webPort!,
                                  _hasWeb,
                                  _hasCli));
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

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      bool isUrl = false}) {
    return Padding(
      padding: const EdgeInsets.only(left: 2, right: 10),
      child: TextFormField(
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          if (isUrl) {
            var splitted = text.split(":");
            if (splitted.length != 2) {
              return "Wrong format for URL: expected host:port";
            }
            if (int.tryParse(splitted[1]) == null) {
              return "Wrong format for URL: port should only have digits";
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
