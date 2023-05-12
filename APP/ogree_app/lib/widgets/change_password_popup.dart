import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';

import '../common/snackbar.dart';

class ChangePasswordPopup extends StatefulWidget {
  @override
  State<ChangePasswordPopup> createState() => _ChangePasswordPopupState();
}

class _ChangePasswordPopupState extends State<ChangePasswordPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isLoading = false;
  String? _oldPassword;
  String? _newPassword;
  String? _confirmPass;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        // height: 240,
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 10),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    "Change password",
                    style: GoogleFonts.inter(
                      fontSize: 22,
                      color: Colors.black,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 20),
                  getFormField(
                      save: (newValue) => _oldPassword = newValue,
                      label: "Current Password"),
                  getFormField(
                      save: (newValue) => _newPassword = newValue,
                      label: "New Password"),
                  getFormField(
                      save: (newValue) => _confirmPass = newValue,
                      label: "Confirm New Password"),
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
                              if (_newPassword != _confirmPass) {
                                showSnackBar(
                                    context, "Password fields do no match",
                                    isError: true);
                                return;
                              }
                              try {
                                setState(() {
                                  _isLoading = true;
                                });
                                var response;
                                response = await changeUserPassword(
                                    _oldPassword!, _newPassword!);
                                if (response == "") {
                                  // widget.parentCallback();
                                  showSnackBar(context, localeMsg.modifyOK,
                                      isSuccess: true);
                                  Navigator.of(context).pop();
                                } else {
                                  setState(() {
                                    _isLoading = false;
                                  });
                                  showSnackBar(context, response,
                                      isError: true);
                                }
                              } catch (e) {
                                showSnackBar(context, e.toString(),
                                    isError: true);
                                return;
                              }
                            }
                          },
                          label: Text(localeMsg.modify),
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
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      String? initial,
      bool isReadOnly = false,
      bool obscure = true}) {
    return Padding(
      padding: const EdgeInsets.only(left: 2, right: 10),
      child: TextFormField(
        obscureText: obscure,
        initialValue: initial,
        readOnly: isReadOnly,
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          return null;
        },
        inputFormatters: formatters,
        decoration: InputDecoration(
          labelText: label,
          prefixText: prefix,
          suffixText: suffix,
        ),
      ),
    );
  }
}
