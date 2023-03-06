import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/language_toggle.dart';

class LoginPage extends StatefulWidget {
  static String tag = 'login-page';

  const LoginPage({super.key});
  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  bool _isChecked = false;
  static const inputStyle = OutlineInputBorder(
    borderSide: BorderSide(
      color: Colors.grey,
      width: 1,
    ),
  );

  String? _email;
  String? _password;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
      body: Container(
        // height: MediaQuery.of(context).size.height,
        decoration: const BoxDecoration(
          image: DecorationImage(
            image: AssetImage("assets/server_background.png"),
            fit: BoxFit.cover,
          ),
        ),
        child: CustomScrollView(slivers: [
          SliverFillRemaining(
            hasScrollBody: false,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Align(
                  alignment: Alignment.topCenter,
                  child: LanguageToggle(),
                ),
                const SizedBox(height: 5),
                Card(
                  // surfaceTintColor: Colors.white,
                  // elevation: 0,
                  child: Form(
                    key: _formKey,
                    child: Container(
                      constraints:
                          const BoxConstraints(maxWidth: 550, maxHeight: 500),
                      padding: const EdgeInsets.only(
                          right: 100, left: 100, top: 50, bottom: 30),
                      child: SingleChildScrollView(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Center(
                                child: Text(localeMsg.welcome,
                                    style: Theme.of(context)
                                        .textTheme
                                        .headlineLarge)),
                            const SizedBox(height: 8),
                            Center(
                              child: Text(
                                localeMsg.welcomeConnect,
                                style:
                                    Theme.of(context).textTheme.headlineMedium,
                              ),
                            ),
                            const SizedBox(height: 25),
                            Center(
                              child: Image.asset(
                                "assets/edf_logo.png",
                                height: 30,
                              ),
                            ),
                            const SizedBox(height: 32),
                            TextFormField(
                              onSaved: (newValue) => _email = newValue,
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: 'E-mail',
                                hintText: 'abc@example.com',
                                labelStyle: GoogleFonts.inter(
                                  fontSize: 11,
                                  color: Colors.black,
                                ),
                                border: inputStyle,
                              ),
                            ),
                            const SizedBox(height: 20),
                            TextFormField(
                              obscureText: true,
                              onSaved: (newValue) => _password = newValue,
                              onEditingComplete: () => tryLogin(),
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: localeMsg.password,
                                hintText: '********',
                                labelStyle: GoogleFonts.inter(
                                  fontSize: 11,
                                  color: Colors.black,
                                ),
                                border: inputStyle,
                              ),
                            ),
                            const SizedBox(height: 25),
                            Wrap(
                              alignment: WrapAlignment.spaceBetween,
                              crossAxisAlignment: WrapCrossAlignment.center,
                              children: [
                                Wrap(
                                  crossAxisAlignment: WrapCrossAlignment.center,
                                  children: [
                                    SizedBox(
                                      height: 24,
                                      width: 24,
                                      child: Checkbox(
                                        value: _isChecked,
                                        onChanged: (bool? value) =>
                                            setState(() => _isChecked = value!),
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      localeMsg.stayLogged,
                                      style: GoogleFonts.inter(
                                        fontSize: 14,
                                        color: Colors.black,
                                      ),
                                    ),
                                  ],
                                ),
                                Text(
                                  localeMsg.forgotPassword,
                                  style: GoogleFonts.inter(
                                    fontSize: 14,
                                    color:
                                        const Color.fromARGB(255, 0, 84, 152),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 30),
                            Align(
                              child: ElevatedButton(
                                onPressed: () => tryLogin(),
                                style: ElevatedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 20,
                                    horizontal: 20,
                                  ),
                                ),
                                child: Text(
                                  localeMsg.login,
                                  style: GoogleFonts.inter(
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
                ),
              ],
            ),
          )
        ]),
      ),
    );
  }

  tryLogin() {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      loginAPI(_email!, _password!)
          .then((value) => value != ""
              ? Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ProjectsPage(
                      userEmail: value,
                    ),
                  ),
                )
              : showSnackBar(
                  context, AppLocalizations.of(context)!.invalidLogin,
                  isError: true))
          .onError((error, stackTrace) => print(error));
    }
  }
}
