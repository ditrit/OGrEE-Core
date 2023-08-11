import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tenant.dart';

class UpdateTenantPopup extends StatefulWidget {
  Function parentCallback;
  Tenant tenant;
  UpdateTenantPopup(
      {super.key, required this.tenant, required this.parentCallback});

  @override
  State<UpdateTenantPopup> createState() => _UpdateTenantPopupState();
}

class _UpdateTenantPopupState extends State<UpdateTenantPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isLoading = false;
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

    return Center(
      child: Container(
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              _isSmallDisplay ? 30 : 40, 20, _isSmallDisplay ? 30 : 40, 15),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Center(
                    child: Text(
                      "Update ${widget.tenant.name}",
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    alignment: WrapAlignment.start,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      // getCheckBox("API", true, (_) {}, enabled: false),
                      getCheckBox(
                          "WEB",
                          widget.tenant.hasWeb,
                          (value) => setState(() {
                                widget.tenant.hasWeb = value!;
                              })),
                      getCheckBox(
                          "DOC",
                          widget.tenant.hasDoc,
                          (value) => setState(() {
                                widget.tenant.hasDoc = value!;
                              })),
                      getCheckBox("Arango", widget.tenant.hasBff, (_) {},
                          enabled: false),
                    ],
                  ),
                  getFormField(
                      save: (newValue) => widget.tenant.imageTag = newValue!,
                      label: localeMsg.deployVersion,
                      icon: Icons.access_time,
                      initial: widget.tenant.imageTag),
                  getFormField(
                    save: (newValue) {
                      var splitted = newValue!.split(":");
                      widget.tenant.apiUrl = splitted[0];
                      widget.tenant.apiPort = splitted[1];
                    },
                    label: "${localeMsg.apiUrl} (hostname:port)",
                    icon: Icons.cloud,
                    prefix: "http://",
                    isUrl: true,
                    initial: "${widget.tenant.apiUrl}:${widget.tenant.apiPort}",
                  ),
                  widget.tenant.hasWeb
                      ? getFormField(
                          save: (newValue) {
                            var splitted = newValue!.split(":");
                            widget.tenant.webUrl = splitted[0];
                            widget.tenant.webPort = splitted[1];
                          },
                          label: "${localeMsg.webUrl} (hostname:port)",
                          icon: Icons.monitor,
                          prefix: "http://",
                          isUrl: true,
                          initial:
                              "${widget.tenant.webUrl}:${widget.tenant.webPort}",
                        )
                      : Container(),
                  widget.tenant.hasDoc
                      ? getFormField(
                          save: (newValue) {
                            var splitted = newValue!.split(":");
                            widget.tenant.docUrl = splitted[0];
                            widget.tenant.docPort = splitted[1];
                          },
                          label: "${localeMsg.docUrl} (hostname:port)",
                          icon: Icons.book,
                          prefix: "http://",
                          isUrl: true,
                          initial:
                              "${widget.tenant.docUrl}:${widget.tenant.docPort}",
                        )
                      : Container(),
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
                          onPressed: () async {
                            if (_formKey.currentState!.validate()) {
                              _formKey.currentState!.save();
                              setState(() {
                                _isLoading = true;
                              });
                              // Create tenant
                              final result = await updateTenant(widget.tenant);
                              switch (result) {
                                case Success():
                                  widget.parentCallback();
                                  showSnackBar(
                                      context, "${localeMsg.modifyOK} 🥳",
                                      isSuccess: true);
                                  Navigator.of(context).pop();
                                case Failure(exception: final exception):
                                  setState(() {
                                    _isLoading = false;
                                  });
                                  showSnackBar(context, exception.toString(),
                                      isError: true);
                              }
                            }
                          },
                          label: const Text("Update"),
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
                              : const Icon(Icons.update_rounded, size: 16))
                    ],
                  )
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  getCheckBox(String title, bool value, Function(bool?) onChange,
      {bool enabled = true}) {
    return SizedBox(
      width: 105,
      child: CheckboxListTile(
        activeColor: Colors.blue.shade600,
        contentPadding: EdgeInsets.zero,
        controlAffinity: ListTileControlAffinity.leading,
        value: value,
        enabled: enabled,
        onChanged: (value) => onChange(value),
        title: Transform.translate(
            offset: const Offset(-10, 0), child: Text(title)),
      ),
    );
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
