import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/common/form_field.dart';

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

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
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
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    localeMsg.changePassword,
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  const SizedBox(height: 20),
                  CustomFormField(
                    save: (newValue) => _oldPassword = newValue,
                    label: localeMsg.currentPassword,
                    icon: Icons.lock_open_rounded,
                  ),
                  CustomFormField(
                    save: (newValue) => _newPassword = newValue,
                    label: localeMsg.newPassword,
                    icon: Icons.lock_outline_rounded,
                  ),
                  CustomFormField(
                    save: (newValue) => _confirmPass = newValue,
                    label: localeMsg.confirmPassword,
                    icon: Icons.lock_outline_rounded,
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
                        onPressed: () => passwordAction(localeMsg),
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

  passwordAction(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      if (_newPassword != _confirmPass) {
        showSnackBar(
          ScaffoldMessenger.of(context),
          localeMsg.passwordNoMatch,
          isError: true,
        );
        return;
      }
      final messenger = ScaffoldMessenger.of(context);
      try {
        setState(() {
          _isLoading = true;
        });
        final response = await changeUserPassword(
          _oldPassword!,
          _newPassword,
        );
        switch (response) {
          case Success():
            showSnackBar(
              messenger,
              localeMsg.modifyOK,
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
