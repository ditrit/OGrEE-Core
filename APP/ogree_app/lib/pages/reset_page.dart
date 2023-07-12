import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/login_page.dart';
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
                                            builder: (context) =>
                                                const LoginPage())),
                                    icon: Icon(
                                      Icons.arrow_back,
                                      color: Colors.blue.shade900,
                                    )),
                                const SizedBox(width: 5),
                                Text(
                                  localeMsg.resetPassword,
                                  style:
                                      Theme.of(context).textTheme.headlineLarge,
                                ),
                              ],
                            ),
                            const SizedBox(height: 25),
                            dotenv.env['ALLOW_SET_BACK'] == "true"
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
                                labelText: localeMsg.newPassword,
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
                              onEditingComplete: () => resetPassword(localeMsg),
                              validator: (text) {
                                if (text == null || text.isEmpty) {
                                  return localeMsg.mandatoryField;
                                }
                                return null;
                              },
                              decoration: InputDecoration(
                                labelText: localeMsg.confirmPassword,
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
                                onPressed: () => resetPassword(localeMsg),
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
                ),
              ],
            ),
          )
        ]),
      ),
    );
  }

  resetPassword(AppLocalizations localeMsg) {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      if (_password != _confirmPassword) {
        showSnackBar(context, localeMsg.passwordNoMatch, isError: true);
        return;
      }
      userResetPassword(_password!, _token!, userUrl: _apiUrl)
          .then((value) => value == ""
              ? resetSucces(localeMsg)
              : showSnackBar(context, value, isError: true))
          .onError((error, stackTrace) {
        print(error);
        showSnackBar(context, error.toString().trim(), isError: true);
      });
    }
  }

  resetSucces(AppLocalizations localeMsg) {
    showSnackBar(context, localeMsg.modifyOK, isSuccess: true);
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => const LoginPage(),
      ),
    );
  }

  backendInput() {
    List<String> options = [];
    if (dotenv.env['BACK_URLS'] != null) {
      options = dotenv.env['BACK_URLS']!.split(",");
    }
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
              labelStyle: const TextStyle(fontSize: 14)),
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
