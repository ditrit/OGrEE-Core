import 'package:flex_color_picker/flex_color_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/widgets/common/form_field.dart';

//Create Nbox: Netbox or Nautobot
class CreateNboxPopup extends StatefulWidget {
  final Function() parentCallback;
  final Tools tool;
  const CreateNboxPopup({
    super.key,
    required this.parentCallback,
    required this.tool,
  });

  @override
  State<CreateNboxPopup> createState() => _CreateNboxPopupState();
}

class _CreateNboxPopupState extends State<CreateNboxPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _userName;
  String? _userPassword;
  String? _port;
  bool _isLoading = false;
  bool _isSmallDisplay = false;
  String netboxVersion = "v4.1-3.0.2";

  @override
  void initState() {
    super.initState();
    _port = widget.tool == Tools.netbox ? "8000" : "8001";
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    final toolName = widget.tool.name.capitalize;
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 360),
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
                    children: [
                      Center(
                        child: Text(
                          "${localeMsg.create} $toolName",
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                      ),
                      const SizedBox(height: 20),
                      Padding(
                        padding: const EdgeInsets.only(bottom: 10, right: 10),
                        child: DropdownButtonFormField<String>(
                          isExpanded: true,
                          borderRadius: BorderRadius.circular(12.0),
                          decoration: GetFormInputDecoration(
                            false,
                            "Version",
                            icon: Icons.bookmark,
                          ),
                          value: netboxVersion,
                          items: const [
                            DropdownMenuItem<String>(
                              value: "v4.1-3.0.2",
                              child: Text(
                                "v4.1-3.0.2",
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            DropdownMenuItem<String>(
                              value: "v3.7-2.8.0",
                              child: Text(
                                "v3.7-2.8.0",
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                          onChanged: (String? value) {
                            // clean the whole form
                            setState(() {
                              netboxVersion = value!;
                            });
                          },
                        ),
                      ),
                      CustomFormField(
                        save: (newValue) => _userName = newValue,
                        label: localeMsg.toolUsername(toolName),
                        icon: Icons.person,
                      ),
                      CustomFormField(
                        save: (newValue) => _userPassword = newValue,
                        label: localeMsg.toolPassword(toolName),
                        icon: Icons.lock,
                      ),
                      CustomFormField(
                        save: (newValue) => _port = newValue,
                        label: localeMsg.toolPort(toolName),
                        initialValue: _port,
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
                            onPressed: () => submitCreateNbox(localeMsg),
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

  submitCreateNbox(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });
      final messenger = ScaffoldMessenger.of(context);
      Result<void, Exception> result;
      if (widget.tool == Tools.netbox) {
        final netbox = Nbox(_userName!, _userPassword!, _port!);
        netbox.version = netboxVersion;
        result = await createNetbox(netbox);
      } else {
        //nautobot
        result = await createNautobot(Nbox(_userName!, _userPassword!, _port!));
      }
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
