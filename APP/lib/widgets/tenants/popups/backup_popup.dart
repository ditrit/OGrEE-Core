import 'dart:convert';
import 'dart:io';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/common/form_field.dart';
import 'package:path_provider/path_provider.dart';
import 'package:universal_html/html.dart' as html;

class BackupPopup extends StatefulWidget {
  final String tenantName;
  const BackupPopup({super.key, required this.tenantName});

  @override
  State<BackupPopup> createState() => _BackupPopupState();
}

class _BackupPopupState extends State<BackupPopup>
    with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  String? _tenantPassword;
  bool _isChecked = false;
  bool _isSmallDisplay = false;
  bool _isLoading = false;
  late TabController _tabController;
  PlatformFile? _loadedFile;
  String? _loadFileResult;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 260),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
            _isSmallDisplay ? 20 : 40,
            8,
            _isSmallDisplay ? 20 : 40,
            15,
          ),
          child: Form(
            key: _formKey,
            child: ScaffoldMessenger(
              child: Builder(
                builder: (context) => Scaffold(
                  backgroundColor: Colors.white,
                  body: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      TabBar(
                        tabAlignment: TabAlignment.center,
                        controller: _tabController,
                        labelStyle: TextStyle(
                          fontSize: 15,
                          fontFamily: GoogleFonts.inter().fontFamily,
                        ),
                        unselectedLabelStyle: TextStyle(
                          fontSize: 15,
                          fontFamily: GoogleFonts.inter().fontFamily,
                        ),
                        isScrollable: true,
                        indicatorSize: TabBarIndicatorSize.label,
                        tabs: [
                          Tab(
                            text: "${localeMsg.toBackup} DB",
                          ),
                          Tab(
                            text: "${localeMsg.restore} DB",
                          ),
                        ],
                      ),
                      SizedBox(
                        height: 189,
                        child: Padding(
                          padding: const EdgeInsets.only(top: 16.0),
                          child: TabBarView(
                            physics: const NeverScrollableScrollPhysics(),
                            controller: _tabController,
                            children: [
                              getBackupView(localeMsg),
                              getRestoreView(localeMsg),
                            ],
                          ),
                        ),
                      ),
                      // const SizedBox(height: 20),
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

  ListView getBackupView(AppLocalizations localeMsg) {
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        const SizedBox(height: 10),
        CustomFormField(
          save: (newValue) => _tenantPassword = newValue,
          label: localeMsg.tenantPassword,
          icon: Icons.lock,
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            const SizedBox(width: 15),
            SizedBox(
              height: 24,
              width: 24,
              child: Checkbox(
                activeColor: Colors.blue.shade600,
                value: _isChecked,
                onChanged: (bool? value) => setState(() => _isChecked = value!),
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
                borderRadius: BorderRadius.all(Radius.circular(12)),
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
                size: 18,
              ),
            ),
          ],
        ),
        const SizedBox(height: 30),
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
              onPressed: requestBackup,
              label: Text(localeMsg.toBackup),
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
                  : const Icon(Icons.history, size: 16),
            ),
          ],
        ),
      ],
    );
  }

  ListView getRestoreView(AppLocalizations localeMsg) {
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        CustomFormField(
          save: (newValue) => _tenantPassword = newValue,
          label: localeMsg.tenantPassword,
          icon: Icons.lock,
        ),
        const SizedBox(height: 2),
        Row(
          children: [
            const SizedBox(width: 15),
            SizedBox(
              height: 24,
              width: 24,
              child: Checkbox(
                activeColor: Colors.blue.shade600,
                value: _isChecked,
                onChanged: (bool? value) => setState(() => _isChecked = value!),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              localeMsg.dropCurrentDB,
              style: const TextStyle(
                fontSize: 14,
                color: Colors.black,
              ),
            ),
            const SizedBox(width: 8),
            Tooltip(
              message: localeMsg.restoreInfoMessage,
              verticalOffset: 13,
              decoration: const BoxDecoration(
                color: Colors.blueAccent,
                borderRadius: BorderRadius.all(Radius.circular(12)),
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
                size: 18,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        if (_loadFileResult == null)
          Align(
            alignment: Alignment.bottomRight,
            child: ElevatedButton.icon(
              onPressed: () async {
                final FilePickerResult? result =
                    await FilePicker.platform.pickFiles(withData: true);
                if (result != null) {
                  setState(() {
                    _loadedFile = result.files.single;
                  });
                }
              },
              icon: Icon(
                _loadedFile != null ? Icons.check_circle : Icons.download,
              ),
              label: _loadedFile != null
                  ? Text(_loadedFile!.name)
                  : Text("${localeMsg.select} ${localeMsg.backup}"),
            ),
          )
        else
          Container(),
        if (_loadFileResult != null)
          Container(
            color: Colors.black,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Text(
                'Result:\n$_loadFileResult',
                style: const TextStyle(color: Colors.white),
              ),
            ),
          )
        else
          Container(),
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
              onPressed: requestRestore,
              label: Text(localeMsg.restore),
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
                  : const Icon(Icons.history, size: 16),
            ),
          ],
        ),
      ],
    );
  }

  requestBackup() async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });
      final localMsg = AppLocalizations.of(context)!;
      final messenger = ScaffoldMessenger.of(context);
      final result =
          await backupTenantDB(widget.tenantName, _tenantPassword!, _isChecked);
      switch (result) {
        case Success(value: final value):
          if (_isChecked) {
            final String filename =
                "${widget.tenantName}_db_${DateFormat('yyyy-MM-ddTHHmmss').format(DateTime.now())}.archive";
            if (kIsWeb) {
              // If web, use html to download csv
              html.AnchorElement(
                href:
                    'data:application/octet-stream;base64,${base64Encode(value)}',
              )
                ..setAttribute("download", filename)
                ..click();
            } else {
              // Save to local filesystem
              final path = (await getApplicationDocumentsDirectory()).path;
              final fileName = '$path/$filename';
              final file = File(fileName);
              file.writeAsBytes(value, flush: true).then(
                    (value) => showSnackBar(
                      messenger,
                      "${localMsg.fileSavedTo} $fileName",
                      copyTextTap: fileName,
                    ),
                  );
            }
          } else {
            showSnackBar(messenger, value, isSuccess: true);
          }
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> requestRestore() async {
    final localeMsg = AppLocalizations.of(context)!;
    if (_loadedFile == null) {
      showSnackBar(ScaffoldMessenger.of(context), localeMsg.mustSelectFile);
      return;
    }
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
        _loadFileResult = null;
      });
      final messenger = ScaffoldMessenger.of(context);
      final result = await restoreTenantDB(
        _loadedFile!,
        widget.tenantName,
        _tenantPassword!,
        _isChecked,
      );
      switch (result) {
        case Success(value: final value):
          showSnackBar(
            messenger,
            "${localeMsg.backupRestored} ${value.substring(value.lastIndexOf("+0000") + 5).trim()}",
            duration: const Duration(seconds: 10),
          );
          setState(() {
            _loadFileResult = value;
            _isLoading = false;
          });
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }
}
