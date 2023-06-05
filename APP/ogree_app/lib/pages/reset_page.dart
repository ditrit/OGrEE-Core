import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/language_toggle.dart';

class ResetPage extends StatefulWidget {
  String token;

  ResetPage({super.key, required this.token});
  @override
  State<ResetPage> createState() => _ResetPageState();
}

class _ResetPageState extends State<ResetPage> {
  final _formKey = GlobalKey<FormState>();
  bool _isChecked = false;
  static const inputStyle = OutlineInputBorder(
    borderSide: BorderSide(
      color: Colors.grey,
      width: 1,
    ),
  );

  String? _token;
  String? _password;
  String? _confirmPassword;
  bool forgot = false;
  String _apiUrl = "";

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    token = widget.token;
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
                  child: Form(
                    key: _formKey,
                    child: Container(
                      constraints:
                          const BoxConstraints(maxWidth: 550, maxHeight: 550),
                      padding: const EdgeInsets.only(
                          right: 100, left: 100, top: 50, bottom: 30),
                      child: SingleChildScrollView(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Row(
                              children: [
                                IconButton(
                                    onPressed: () => Navigator.of(context).push(
                                        MaterialPageRoute(
                                            builder: (context) => LoginPage())),
                                    icon: Icon(
                                      Icons.arrow_back,
                                      color: Colors.blue.shade900,
                                    )),
                                const SizedBox(width: 5),
                                Text(
                                  "Reset password",
                                  style:
                                      Theme.of(context).textTheme.headlineLarge,
                                ),
                              ],
                            ),
                            const SizedBox(height: 25),
                            allowBackChoice
                                ? backendInput()
                                : Center(
                                    child: Image.asset(
                                      "assets/edf_logo.png",
                                      height: 30,
                                    ),
                                  ),
                            const SizedBox(height: 32),
                            TextFormField(
                              initialValue: widget.token,
                              enabled: widget.token == "",
                              onSaved: (newValue) => _token = newValue,
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: 'Reset Token',
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
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: 'New password',
                                hintText: '********',
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
                              onSaved: (newValue) =>
                                  _confirmPassword = newValue,
                              onEditingComplete: () => resetPassword(),
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: "Confirm new password",
                                hintText: '********',
                                labelStyle: GoogleFonts.inter(
                                  fontSize: 11,
                                  color: Colors.black,
                                ),
                                border: inputStyle,
                              ),
                            ),
                            const SizedBox(height: 25),
                            Align(
                              child: ElevatedButton(
                                onPressed: () => resetPassword(),
                                style: ElevatedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 20,
                                    horizontal: 20,
                                  ),
                                ),
                                child: Text(
                                  "Reset",
                                  style: TextStyle(
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

  resetPassword() {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      if (_password != _confirmPassword) {
        showSnackBar(context, "Password fields do no match", isError: true);
        return;
      }
      userResetPassword(_password!, _token!, userUrl: _apiUrl)
          .then((value) => value == ""
              ? resetSucces()
              : showSnackBar(context, value, isError: true))
          .onError((error, stackTrace) {
        print(error);
        showSnackBar(context, error.toString().trim(), isError: true);
      });
    }
  }

  resetSucces() {
    showSnackBar(context, "Password successfully changed", isSuccess: true);
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => LoginPage(),
      ),
    );
  }

  backendInput() {
    final options = backendUrl.split(",");
    final localeMsg = AppLocalizations.of(context)!;
    return RawAutocomplete<String>(
      optionsBuilder: (TextEditingValue textEditingValue) {
        return options.where((String option) {
          return option.contains(textEditingValue.text);
        });
      },
      fieldViewBuilder: (BuildContext context,
          TextEditingController textEditingController,
          FocusNode focusNode,
          VoidCallback onFieldSubmitted) {
        textEditingController.text = options.first;
        return TextFormField(
          controller: textEditingController,
          focusNode: focusNode,
          onSaved: (newValue) => _apiUrl = newValue!,
          validator: (text) {
            if (text == null || text.trim().isEmpty) {
              return localeMsg.mandatoryField;
            }
            return null;
          },
          decoration: InputDecoration(
              isDense: true,
              labelText: localeMsg.selectServer,
              labelStyle: TextStyle(fontSize: 14)),
          onTap: () {
            textEditingController.clear();
          },
        );
      },
      optionsViewBuilder: (BuildContext context,
          AutocompleteOnSelected<String> onSelected, Iterable<String> options) {
        return Align(
          alignment: Alignment.topLeft,
          child: Material(
            elevation: 4.0,
            child: SizedBox(
              height: options.length > 2 ? 171.0 : 57.0 * options.length,
              width: 350,
              child: ListView.builder(
                padding: const EdgeInsets.all(8.0),
                itemCount: options.length,
                itemBuilder: (BuildContext context, int index) {
                  final String option = options.elementAt(index);
                  return GestureDetector(
                    onTap: () {
                      onSelected(option);
                    },
                    child: ListTile(
                      title: Text(option, style: const TextStyle(fontSize: 14)),
                    ),
                  );
                },
              ),
            ),
          ),
        );
      },
    );
  }
}

String backendUrl = const String.fromEnvironment(
  'BACK_URLS',
  defaultValue: 'http://localhost:3001',
);

bool allowBackChoice = const bool.fromEnvironment(
  'ALLOW_SET_BACK',
  defaultValue: true,
);
