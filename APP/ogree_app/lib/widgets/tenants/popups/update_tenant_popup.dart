import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
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

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        // height: 240,
        width: 500,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.add_to_photos),
                      Text(
                        "   Update " + widget.tenant.name,
                        style: GoogleFonts.inter(
                          fontSize: 22,
                          color: Colors.black,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 40),
                  Wrap(
                    alignment: WrapAlignment.start,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Padding(
                        padding: const EdgeInsets.only(right: 12),
                        child: Text("Services:"),
                      ),
                      getCheckBox("API", true, (_) {}, enabled: false),
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
                    ],
                  ),
                  getFormField(
                      save: (newValue) => widget.tenant.imageTag = newValue!,
                      label: "Version du déploiement (branch)",
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
                  const SizedBox(height: 30),
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
                              // Load logo first, if provided
                              String response = localeMsg.notLoaded;
                              // Create tenant
                              response = await updateTenant(widget.tenant);
                              if (response == "") {
                                widget.parentCallback();
                                showSnackBar(
                                    context, "Tenant successfully updated 🥳",
                                    isSuccess: true);
                                Navigator.of(context).pop();
                              } else {
                                setState(() {
                                  _isLoading = false;
                                });
                                showSnackBar(context, response, isError: true);
                              }
                            }
                          },
                          label: Text("Update"),
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
      width: 100,
      child: CheckboxListTile(
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
      padding: const EdgeInsets.only(left: 2, right: 10),
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
        decoration: InputDecoration(
          icon: Icon(icon, color: Colors.blue.shade900),
          labelText: label,
          prefixText: prefix,
          suffixText: suffix,
        ),
      ),
    );
  }
}
