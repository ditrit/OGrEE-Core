import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/common/form_field.dart';

class CreateOpenDcimPopup extends StatefulWidget {
  final Function() parentCallback;
  const CreateOpenDcimPopup({super.key, required this.parentCallback});

  @override
  State<CreateOpenDcimPopup> createState() => _CreateOpenDcimPopupState();
}

class _CreateOpenDcimPopupState extends State<CreateOpenDcimPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _dcimPort = "80";
  String? _adminerPort = "8080";
  bool _isLoading = false;
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 250),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
            _isSmallDisplay ? 30 : 40,
            20,
            _isSmallDisplay ? 30 : 40,
            15,
          ),
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
                          "${localeMsg.create} OpenDCIM",
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                      ),
                      // const Divider(height: 35),
                      const SizedBox(height: 20),
                      CustomFormField(
                        save: (newValue) => _dcimPort = newValue,
                        label: "OpenDCIM port",
                        initialValue: _dcimPort,
                        icon: Icons.numbers,
                        formatters: <TextInputFormatter>[
                          FilteringTextInputFormatter.digitsOnly,
                          LengthLimitingTextInputFormatter(4),
                        ],
                      ),
                      CustomFormField(
                        save: (newValue) => _adminerPort = newValue,
                        label: "Adminer port",
                        initialValue: _adminerPort,
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
                            onPressed: () => submitCreateOpenDcim(localeMsg),
                            label: Text(localeMsg.create),
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
                                : const Icon(
                                    Icons.check_circle,
                                    size: 16,
                                  ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  submitCreateOpenDcim(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });
      final messenger = ScaffoldMessenger.of(context);
      // Create dcim
      final result = await createOpenDcim(_dcimPort!, _adminerPort);
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, "${localeMsg.createOK} 🥳", isSuccess: true);
          if (mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }
}
