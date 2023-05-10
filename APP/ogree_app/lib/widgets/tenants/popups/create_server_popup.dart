import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/models/tenant.dart';

class CreateServerPopup extends StatefulWidget {
  Function() parentCallback;
  CreateServerPopup({super.key, required this.parentCallback});

  @override
  State<CreateServerPopup> createState() => _CreateServerPopupState();
}

enum AuthOption { pKey, password }

class _CreateServerPopupState extends State<CreateServerPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _sshHost;
  String? _sshUser;
  String? _sshKey;
  String? _sshKeyPass;
  String? _sshPassword;
  String? _installPath;
  String? _port;
  bool _isLoading = false;
  AuthOption? _authOption = AuthOption.pKey;

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
                        "   ${localeMsg.createServer}",
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
                      save: (newValue) => _sshHost = newValue,
                      label: "SSH Host",
                      icon: Icons.dns),
                  getFormField(
                      save: (newValue) => _sshUser = newValue,
                      label: "SSH User",
                      icon: Icons.person),
                  SizedBox(height: 8),
                  Wrap(
                    children: <Widget>[
                      SizedBox(
                        width: 200,
                        child: RadioListTile<AuthOption>(
                          dense: true,
                          title: const Text('Private Key'),
                          value: AuthOption.pKey,
                          groupValue: _authOption,
                          onChanged: (AuthOption? value) {
                            setState(() {
                              _authOption = value;
                            });
                          },
                        ),
                      ),
                      SizedBox(
                        width: 200,
                        child: RadioListTile<AuthOption>(
                          dense: true,
                          title: Text(localeMsg.password),
                          value: AuthOption.password,
                          groupValue: _authOption,
                          onChanged: (AuthOption? value) {
                            setState(() {
                              _authOption = value;
                            });
                          },
                        ),
                      ),
                    ],
                  ),
                  _authOption == AuthOption.pKey
                      ? Column(
                          children: [
                            getFormField(
                                save: (newValue) => _sshKey = newValue,
                                label: "SSH Private Key (/local/path/file)",
                                icon: Icons.lock),
                            getFormField(
                                save: (newValue) => _sshKeyPass = newValue,
                                label:
                                    "Private Key Passphrase (${localeMsg.optional})",
                                icon: Icons.lock,
                                shouldValidate: false),
                          ],
                        )
                      : getFormField(
                          save: (newValue) => _sshPassword = newValue,
                          label: localeMsg.password,
                          icon: Icons.lock),
                  getFormField(
                      save: (newValue) => _installPath = newValue,
                      label: localeMsg.serverPath,
                      icon: Icons.folder),
                  getFormField(
                      save: (newValue) => _port = newValue,
                      label: localeMsg.portServer,
                      icon: Icons.onetwothree,
                      formatters: [FilteringTextInputFormatter.digitsOnly]),
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

                              var response = _authOption == AuthOption.pKey
                                  ? await createBackendServer(<String, String>{
                                      'host': _sshHost!,
                                      'user': _sshUser!,
                                      'pkey': _sshKey!,
                                      'pkeypass': _sshKeyPass.toString(),
                                      'dstpath': _installPath!,
                                      'runport': _port!,
                                    })
                                  : await createBackendServer(<String, String>{
                                      'host': _sshHost!,
                                      'user': _sshUser!,
                                      'password': _sshPassword!,
                                      'dstpath': _installPath!,
                                      'runport': _port!,
                                    });
                              if (response == "") {
                                widget.parentCallback();
                                showSnackBar(context, localeMsg.createOK,
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
      bool shouldValidate = true}) {
    return Padding(
      padding: const EdgeInsets.only(left: 2, right: 10),
      child: TextFormField(
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (shouldValidate) {
            if (text == null || text.isEmpty) {
              return AppLocalizations.of(context)!.mandatoryField;
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
