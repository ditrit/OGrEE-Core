import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:ogree_app/widgets/common/form_field.dart';
import 'package:universal_html/html.dart' as html;
import 'dart:convert';
import 'dart:io';
import 'package:path_provider/path_provider.dart';

class BackupPopup extends StatefulWidget {
  String tenantName;
  BackupPopup({super.key, required this.tenantName});

  @override
  State<BackupPopup> createState() => _BackupPopupState();
}

class _BackupPopupState extends State<BackupPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _tenantPassword;
  bool _shouldDownload = false;
  bool _isSmallDisplay = false;
  bool _isLoading = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 230),
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
                            children: [
                              Center(
                                  child: Text(
                                localeMsg.requestBackup,
                                style:
                                    Theme.of(context).textTheme.headlineMedium,
                              )),
                              const SizedBox(height: 20),
                              CustomFormField(
                                  save: (newValue) =>
                                      _tenantPassword = newValue,
                                  label: localeMsg.tenantPassword,
                                  icon: Icons.lock),
                              const SizedBox(height: 10),
                              Row(
                                children: [
                                  const SizedBox(width: 15),
                                  SizedBox(
                                    height: 24,
                                    width: 24,
                                    child: Checkbox(
                                      activeColor: Colors.blue.shade600,
                                      value: _shouldDownload,
                                      onChanged: (bool? value) => setState(
                                          () => _shouldDownload = value!),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    localeMsg.downloadBackup,
                                    style: const TextStyle(
                                      fontSize: 14,
                                      color: Colors.black,
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Tooltip(
                                    message: localeMsg.backupInfoMessage,
                                    verticalOffset: 13,
                                    decoration: const BoxDecoration(
                                      color: Colors.blueAccent,
                                      borderRadius:
                                          BorderRadius.all(Radius.circular(12)),
                                    ),
                                    textStyle: const TextStyle(
                                      fontSize: 13,
                                      color: Colors.white,
                                      height: 1.5,
                                    ),
                                    padding: const EdgeInsets.all(16),
                                    child: const Icon(
                                        Icons.info_outline_rounded,
                                        color: Colors.blueAccent,
                                        size: 18),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 20),
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
                                      onPressed: requestBackup,
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

  requestBackup() async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      final messenger = ScaffoldMessenger.of(context);
      final result = await backupTenantDB(
          widget.tenantName, _tenantPassword!, _shouldDownload);
      switch (result) {
        case Success(value: final value):
          if (_shouldDownload) {
            String filename =
                "${widget.tenantName}_db_${DateFormat('yyyy-MM-ddTHHmmss').format(DateTime.now())}.archive";
            if (kIsWeb) {
              // If web, use html to download csv
              html.AnchorElement(
                  href:
                      'data:application/octet-stream;base64,${base64Encode(value)}')
                ..setAttribute("download", filename)
                ..click();
            } else {
              // Save to local filesystem
              var path = (await getApplicationDocumentsDirectory()).path;
              var fileName = '$path/$filename';
              var file = File(fileName);
              file.writeAsBytes(value, flush: true).then((value) =>
                  showSnackBar(messenger,
                      "${AppLocalizations.of(context)!.fileSavedTo} $fileName",
                      copyTextTap: fileName));
            }
          } else {
            showSnackBar(messenger, value, isSuccess: true);
          }
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }
}
