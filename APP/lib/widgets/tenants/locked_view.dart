// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tenant.dart';

class LockedView extends StatefulWidget {
  final Tenant tenant;
  final Function parentCallback;
  const LockedView({
    super.key,
    required this.tenant,
    required this.parentCallback,
  });
  @override
  State<LockedView> createState() => _LockedViewState();
}

class _LockedViewState extends State<LockedView> {
  String? _email;
  String? _password;
  final formKey = GlobalKey<FormState>();

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final bool isSmallDisplay =
        IsSmallDisplay(MediaQuery.of(context).size.width);
    return Form(
      key: formKey,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 350, maxHeight: 500),
        padding: EdgeInsets.only(
          right: isSmallDisplay ? 32 : 100,
          left: isSmallDisplay ? 15 : 100,
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.lock, size: isSmallDisplay ? 32 : 64),
            const SizedBox(height: 20),
            Text(localeMsg.loginTenant),
            const SizedBox(height: 20),
            SizedBox(
              width: 350,
              child: TextFormField(
                onSaved: (newValue) => _email = newValue,
                validator: (text) {
                  if (text == null || text.isEmpty) {
                    return localeMsg.mandatoryField;
                  }
                  return null;
                },
                decoration: GetFormInputDecoration(
                  isSmallDisplay,
                  'E-mail',
                  icon: Icons.alternate_email,
                  hint: 'abc@example.com',
                ),
              ),
            ),
            const SizedBox(height: 20),
            Container(
              constraints: const BoxConstraints(maxWidth: 350),
              child: TextFormField(
                obscureText: true,
                onSaved: (newValue) => _password = newValue,
                onEditingComplete: () => tryLogin(formKey),
                validator: (text) {
                  if (text == null || text.isEmpty) {
                    return localeMsg.mandatoryField;
                  }
                  return null;
                },
                decoration: GetFormInputDecoration(
                  isSmallDisplay,
                  localeMsg.password,
                  icon: Icons.lock_outline_rounded,
                  hint: '********',
                ),
              ),
            ),
            const SizedBox(height: 20),
            Align(
              child: ElevatedButton(
                onPressed: () => tryLogin(formKey),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    vertical: 20,
                    horizontal: 20,
                  ),
                ),
                child: Text(
                  localeMsg.login,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  tryLogin(formKey) async {
    if (formKey.currentState!.validate()) {
      formKey.currentState!.save();
      final messenger = ScaffoldMessenger.of(context);
      final localeMsg = AppLocalizations.of(context)!;
      final result = await loginAPITenant(
        _email!,
        _password!,
        "${widget.tenant.apiUrl}:${widget.tenant.apiPort}",
      );
      switch (result) {
        case Success():
          widget.parentCallback();
        case Failure(exception: final exception):
          final String errorMsg = exception.toString() == "Exception"
              ? localeMsg.invalidLogin
              : exception.toString();
          showSnackBar(messenger, errorMsg, isError: true);
      }
    }
  }
}
