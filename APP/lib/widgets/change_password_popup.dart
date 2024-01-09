import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/theme.dart';

import '../common/snackbar.dart';

class ChangePasswordPopup extends StatefulWidget {
  const ChangePasswordPopup({super.key});

  @override
  State<ChangePasswordPopup> createState() => _ChangePasswordPopupState();
}

class _ChangePasswordPopupState extends State<ChangePasswordPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isLoading = false;
  String? _oldPassword;
  String? _newPassword;
  String? _confirmPass;
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        // height: 240,
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 10),
        decoration: PopupDecoration,
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
                    localeMsg.changePassword,
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  const SizedBox(height: 20),
                  getFormField(
                      save: (newValue) => _oldPassword = newValue,
                      label: localeMsg.currentPassword,
                      icon: Icons.lock_open_rounded),
                  getFormField(
                      save: (newValue) => _newPassword = newValue,
                      label: localeMsg.newPassword,
                      icon: Icons.lock_outline_rounded),
                  getFormField(
                      save: (newValue) => _confirmPass = newValue,
                      label: localeMsg.confirmPassword,
                      icon: Icons.lock_outline_rounded),
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
                              if (_newPassword != _confirmPass) {
                                showSnackBar(ScaffoldMessenger.of(context),
                                    localeMsg.passwordNoMatch,
                                    isError: true);
                                return;
                              }
                              final messenger = ScaffoldMessenger.of(context);
                              try {
                                setState(() {
                                  _isLoading = true;
                                });
                                final response = await changeUserPassword(
                                    _oldPassword!, _newPassword!);
                                switch (response) {
                                  case Success():
                                    showSnackBar(messenger, localeMsg.modifyOK,
                                        isSuccess: true);
                                    if (context.mounted) {
                                      Navigator.of(context).pop();
                                    }
                                  case Failure(exception: final exception):
                                    setState(() {
                                      _isLoading = false;
                                    });
                                    showSnackBar(
                                        messenger, exception.toString(),
                                        isError: true);
                                }
                              } catch (e) {
                                showSnackBar(messenger, e.toString(),
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
      required IconData icon,
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      String? initial,
      bool isReadOnly = false,
      bool obscure = true}) {
    return Padding(
      padding: FormInputPadding,
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
        decoration: GetFormInputDecoration(_isSmallDisplay, label,
            prefixText: prefix, suffixText: suffix, icon: icon),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
