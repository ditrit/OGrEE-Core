import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/netbox.dart';

class CreateNetboxPopup extends StatefulWidget {
  Function() parentCallback;
  CreateNetboxPopup({super.key, required this.parentCallback});

  @override
  State<CreateNetboxPopup> createState() => _CreateNetboxPopupState();
}

class _CreateNetboxPopupState extends State<CreateNetboxPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _userName;
  String? _userPassword;
  String? _port = "8000";
  bool _isLoading = false;
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 300),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              _isSmallDisplay ? 30 : 40, 20, _isSmallDisplay ? 30 : 40, 15),
          child: Form(
            key: _formKey,
            child: ScaffoldMessenger(
                child: Builder(
                    builder: (context) => Scaffold(
                          backgroundColor: Colors.white,
                          body: ListView(
                            padding: EdgeInsets.zero,
                            //shrinkWrap: true,
                            children: [
                              Center(
                                  child: Text(
                                "Créer netbox",
                                style:
                                    Theme.of(context).textTheme.headlineMedium,
                              )),
                              // const Divider(height: 35),
                              const SizedBox(height: 20),
                              getFormField(
                                  save: (newValue) => _userName = newValue,
                                  label: "Netbox user name",
                                  icon: Icons.person),
                              getFormField(
                                  save: (newValue) => _userPassword = newValue,
                                  label: "Netbox user password",
                                  icon: Icons.lock),
                              getFormField(
                                save: (newValue) => _port = newValue,
                                label: "Netbox port",
                                initial: _port,
                                icon: Icons.numbers,
                                formatters: <TextInputFormatter>[
                                  FilteringTextInputFormatter.digitsOnly,
                                  LengthLimitingTextInputFormatter(4),
                                ],
                              ),
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
                                      onPressed: () =>
                                          submitCreateNetbox(localeMsg),
                                      label: Text(localeMsg.create),
                                      icon: _isLoading
                                          ? Container(
                                              width: 24,
                                              height: 24,
                                              padding:
                                                  const EdgeInsets.all(2.0),
                                              child:
                                                  const CircularProgressIndicator(
                                                color: Colors.white,
                                                strokeWidth: 3,
                                              ),
                                            )
                                          : const Icon(Icons.check_circle,
                                              size: 16))
                                ],
                              )
                            ],
                          ),
                        ))),
          ),
        ),
      ),
    );
  }

  submitCreateNetbox(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });
      // Create tenant
      var result =
          await createNetbox(Netbox(_userName!, _userPassword!, _port!));
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(context, "${localeMsg.createOK} 🥳", isSuccess: true);
          Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(context, exception.toString(), isError: true);
      }
    }
  }

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      String? initial,
      bool isUrl = false}) {
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
        initialValue: initial,
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          if (isUrl) {
            var splitted = text.split(":");
            if (splitted.length != 2) {
              return AppLocalizations.of(context)!.wrongFormatUrl;
            }
            if (int.tryParse(splitted[1]) == null) {
              return AppLocalizations.of(context)!.wrongFormatPort;
            }
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
