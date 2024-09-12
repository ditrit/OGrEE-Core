import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/pages/projects_page.dart';

class LoginCard extends StatefulWidget {
  const LoginCard({super.key});

  @override
  State<LoginCard> createState() => _LoginCardState();
}

class _LoginCardState extends State<LoginCard> {
  final _formKey = GlobalKey<FormState>();
  bool _stayLoggedIn = false;
  String? _email;
  String? _password;
  String _apiUrl = "";
  BackendType? apiType;

  bool showForgotView = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final bool isSmallDisplay =
        IsSmallDisplay(MediaQuery.of(context).size.width);
    return Card(
      child: Form(
        key: _formKey,
        child: Container(
          constraints: const BoxConstraints(maxWidth: 550, maxHeight: 520),
          padding: EdgeInsets.only(
            right: isSmallDisplay ? 45 : 100,
            left: isSmallDisplay ? 45 : 100,
            top: 50,
            bottom: 30,
          ),
          child: SingleChildScrollView(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (showForgotView)
                  Wrap(
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      IconButton(
                        constraints: const BoxConstraints(),
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
                      SizedBox(width: isSmallDisplay ? 0 : 5),
                      Text(
                        localeMsg.resetPassword,
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                    ],
                  )
                else
                  Center(
                    child: Text(
                      localeMsg.welcome,
                      style: Theme.of(context).textTheme.headlineLarge,
                    ),
                  ),
                const SizedBox(height: 8),
                if (showForgotView)
                  const SizedBox(height: 10)
                else
                  Center(
                    child: Text(
                      localeMsg.welcomeConnect,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                  ),
                if (showForgotView) Container() else const SizedBox(height: 20),
                if (dotenv.env['ALLOW_SET_BACK'] == "true")
                  BackendInput(
                    parentCallback: (newValue) => _apiUrl = newValue,
                  )
                else
                  Center(
                    child: Image.asset(
                      "assets/custom/logo.png",
                      height: 40,
                    ),
                  ),
                if (dotenv.env['ALLOW_SET_BACK'] == "true")
                  Align(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      child: Badge(
                        backgroundColor: Colors.white,
                        label: Text(
                          getBackendTypeText(),
                          style: const TextStyle(color: Colors.black),
                        ),
                      ),
                    ),
                  )
                else
                  const SizedBox(height: 30),
                TextFormField(
                  onSaved: (newValue) => _email = newValue,
                  validator: (text) {
                    if (text == null || text.isEmpty) {
                      return localeMsg.mandatoryField;
                    }
                    return null;
                  },
                  decoration: LoginInputDecoration(
                    label: 'E-mail',
                    hint: 'abc@example.com',
                    isSmallDisplay: isSmallDisplay,
                  ),
                ),
                SizedBox(height: isSmallDisplay ? 10 : 20),
                if (showForgotView)
                  Container()
                else
                  TextFormField(
                    obscureText: true,
                    onSaved: (newValue) => _password = newValue,
                    onEditingComplete: () =>
                        tryLogin(localeMsg, ScaffoldMessenger.of(context)),
                    validator: (text) {
                      if (!showForgotView && (text == null || text.isEmpty)) {
                        return localeMsg.mandatoryField;
                      }
                      return null;
                    },
                    decoration: LoginInputDecoration(
                      label: localeMsg.password,
                      hint: '********',
                      isSmallDisplay: isSmallDisplay,
                    ),
                  ),
                if (!showForgotView)
                  SizedBox(height: isSmallDisplay ? 15 : 25)
                else
                  Container(),
                if (showForgotView)
                  TextButton(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (context) => const LoginPage(
                          isPasswordReset: true,
                        ),
                      ),
                    ),
                    child: Text(
                      localeMsg.haveResetToken,
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color.fromARGB(255, 0, 84, 152),
                      ),
                    ),
                  )
                else
                  Wrap(
                    alignment: WrapAlignment.spaceBetween,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      if (!isSmallDisplay)
                        Wrap(
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            SizedBox(
                              height: 24,
                              width: 24,
                              child: StatefulBuilder(
                                builder: (context, localSetState) {
                                  return Checkbox(
                                    value: _stayLoggedIn,
                                    onChanged: (bool? value) => localSetState(
                                      () => _stayLoggedIn = value!,
                                    ),
                                  );
                                },
                              ),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              localeMsg.stayLogged,
                              style: const TextStyle(
                                fontSize: 14,
                                color: Colors.black,
                              ),
                            ),
                          ],
                        )
                      else
                        Container(),
                      TextButton(
                        onPressed: () => setState(() {
                          showForgotView = !showForgotView;
                        }),
                        child: Text(
                          localeMsg.forgotPassword,
                          style: const TextStyle(
                            fontSize: 14,
                            color: Color.fromARGB(255, 0, 84, 152),
                          ),
                        ),
                      ),
                    ],
                  ),
                SizedBox(
                  height: showForgotView ? 20 : (isSmallDisplay ? 15 : 30),
                ),
                Align(
                  child: ElevatedButton(
                    onPressed: () => showForgotView
                        ? resetPassword(
                            localeMsg,
                            ScaffoldMessenger.of(context),
                          )
                        : tryLogin(localeMsg, ScaffoldMessenger.of(context)),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        vertical: 20,
                        horizontal: 20,
                      ),
                    ),
                    child: Text(
                      showForgotView ? localeMsg.reset : localeMsg.login,
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

  resetPassword(
    AppLocalizations localeMsg,
    ScaffoldMessengerState messenger,
  ) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      final result = await userForgotPassword(_email!, userUrl: _apiUrl);
      switch (result) {
        case Success():
          showSnackBar(messenger, localeMsg.resetSent, isSuccess: true);
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString().trim(), isError: true);
      }
    }
  }

  tryLogin(AppLocalizations localeMsg, ScaffoldMessengerState messenger) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      final result = await loginAPI(
        _email!,
        _password!,
        userUrl: _apiUrl,
        stayLoggedIn: _stayLoggedIn,
      );
      switch (result) {
        case Success(value: final loginData):
          if (apiType == BackendType.tenant) {
            await fetchApiVersion(_apiUrl);
          }
          if (mounted) {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => ProjectsPage(
                  userEmail: loginData.first,
                  isTenantMode: loginData[1] == "true",
                ),
              ),
            );
          }
        case Failure(exception: final exception):
          final String errorMsg = exception.toString() == "Exception"
              ? localeMsg.invalidLogin
              : exception.toString();
          showSnackBar(messenger, errorMsg, isError: true);
      }
    }
  }

  getBackendType(inputUrl) async {
    final result = await fetchApiVersion(inputUrl);
    switch (result) {
      case Success(value: final type):
        setState(() {
          apiType = type;
        });
      case Failure(exception: final exception):
        print(exception);
        setState(() {
          apiType = BackendType.unavailable;
        });
    }
  }

  String getBackendTypeText() {
    if (apiType == null) {
      return "";
    } else if (apiType == BackendType.unavailable) {
      return AppLocalizations.of(context)!.unavailable.toUpperCase();
    } else {
      return "${apiType!.name.toUpperCase()} SERVER";
    }
  }
}

class BackendInput extends StatelessWidget {
  final Function(String) parentCallback;
  const BackendInput({super.key, required this.parentCallback});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    List<String> options = [];
    if (dotenv.env['BACK_URLS'] != null) {
      options = dotenv.env['BACK_URLS']!.split(",");
    }
    return RawAutocomplete<String>(
      optionsBuilder: (TextEditingValue textEditingValue) {
        return options.where((String option) {
          return option.contains(textEditingValue.text);
        });
      },
      fieldViewBuilder: (
        BuildContext context,
        TextEditingController textEditingController,
        FocusNode focusNode,
        VoidCallback onFieldSubmitted,
      ) {
        textEditingController.text = options.first;
        return TextFormField(
          controller: textEditingController,
          focusNode: focusNode,
          onSaved: (newValue) => parentCallback(newValue!),
          validator: (text) {
            if (text == null || text.trim().isEmpty) {
              return localeMsg.mandatoryField;
            }
            return null;
          },
          decoration: InputDecoration(
            isDense: true,
            labelText: localeMsg.selectServer,
            labelStyle: const TextStyle(fontSize: 14),
          ),
        );
      },
      optionsViewBuilder: (
        BuildContext context,
        AutocompleteOnSelected<String> onSelected,
        Iterable<String> options,
      ) {
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
