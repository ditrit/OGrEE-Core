import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/widgets/login/login_card.dart';

class ResetCard extends StatelessWidget {
  final _formKey = GlobalKey<FormState>();
  String token;
  String? _token;
  String? _password;
  String? _confirmPassword;
  String _apiUrl = "";

  ResetCard({super.key, required this.token});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final bool isSmallDisplay =
        IsSmallDisplay(MediaQuery.of(context).size.width);
    return Card(
      child: Form(
        key: _formKey,
        child: Container(
          constraints: const BoxConstraints(maxWidth: 550, maxHeight: 550),
          padding:
              const EdgeInsets.only(right: 100, left: 100, top: 50, bottom: 30),
          child: SingleChildScrollView(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    IconButton(
                      onPressed: () => Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => const LoginPage(),
                        ),
                      ),
                      icon: Icon(
                        Icons.arrow_back,
                        color: Colors.blue.shade900,
                      ),
                    ),
                    const SizedBox(width: 5),
                    Text(
                      localeMsg.resetPassword,
                      style: Theme.of(context).textTheme.headlineLarge,
                    ),
                  ],
                ),
                const SizedBox(height: 25),
                if (dotenv.env['ALLOW_SET_BACK'] == "true")
                  BackendInput(
                    parentCallback: (newValue) => _apiUrl = newValue,
                  )
                else
                  Center(
                    child: Image.asset(
                      "assets/custom/logo.png",
                      height: 30,
                    ),
                  ),
                const SizedBox(height: 32),
                TextFormField(
                  initialValue: token,
                  enabled: token == "",
                  onSaved: (newValue) => _token = newValue,
                  validator: (text) {
                    if (text == null || text.isEmpty) {
                      return localeMsg.mandatoryField;
                    }
                    return null;
                  },
                  decoration: LoginInputDecoration(
                    label: 'Reset Token',
                    isSmallDisplay: isSmallDisplay,
                  ),
                ),
                const SizedBox(height: 20),
                TextFormField(
                  obscureText: true,
                  onSaved: (newValue) => _password = newValue,
                  validator: (text) {
                    if (text == null || text.isEmpty) {
                      return localeMsg.mandatoryField;
                    }
                    return null;
                  },
                  decoration: LoginInputDecoration(
                    label: localeMsg.newPassword,
                    hint: '********',
                    isSmallDisplay: isSmallDisplay,
                  ),
                ),
                const SizedBox(height: 20),
                TextFormField(
                  obscureText: true,
                  onSaved: (newValue) => _confirmPassword = newValue,
                  onEditingComplete: () => resetPassword(localeMsg, context),
                  validator: (text) {
                    if (text == null || text.isEmpty) {
                      return localeMsg.mandatoryField;
                    }
                    return null;
                  },
                  decoration: LoginInputDecoration(
                    label: localeMsg.confirmPassword,
                    hint: '********',
                    isSmallDisplay: isSmallDisplay,
                  ),
                ),
                const SizedBox(height: 25),
                Align(
                  child: ElevatedButton(
                    onPressed: () => resetPassword(localeMsg, context),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        vertical: 20,
                        horizontal: 20,
                      ),
                    ),
                    child: Text(
                      localeMsg.reset,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 15),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> resetPassword(
      AppLocalizations localeMsg, BuildContext context) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      if (_password != _confirmPassword) {
        showSnackBar(
          ScaffoldMessenger.of(context),
          localeMsg.passwordNoMatch,
          isError: true,
        );
        return;
      }
      final messenger = ScaffoldMessenger.of(context);
      final result =
          await userResetPassword(_password!, _token!, userUrl: _apiUrl);
      switch (result) {
        case Success():
          resetSuccess(localeMsg, context);
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString().trim(), isError: true);
      }
    }
  }

  resetSuccess(AppLocalizations localeMsg, BuildContext context) {
    showSnackBar(
      ScaffoldMessenger.of(context),
      localeMsg.modifyOK,
      isSuccess: true,
    );
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => const LoginPage(),
      ),
    );
  }
}
