import 'dart:io';

import 'package:flutter/material.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:path_provider/path_provider.dart';
import 'package:universal_html/html.dart' as html;
import 'package:http/http.dart' as http;

enum CliOS { windows, linux, macOS }

class DownloadCliPopup extends StatefulWidget {
  const DownloadCliPopup({super.key});

  @override
  State<DownloadCliPopup> createState() => _DownloadCliPopupState();
}

class _DownloadCliPopupState extends State<DownloadCliPopup> {
  CliOS _selectedOS = CliOS.windows;
  bool _isLoading = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 400,
        constraints: const BoxConstraints(maxHeight: 190),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              isSmallDisplay ? 30 : 40, 20, isSmallDisplay ? 30 : 40, 15),
          child: ScaffoldMessenger(
              child: Builder(
                  builder: (context) => Scaffold(
                        backgroundColor: Colors.white,
                        body: ListView(
                          padding: EdgeInsets.zero,
                          children: [
                            Center(
                                child: Text(
                              localeMsg.downloadCliTitle,
                              style: Theme.of(context).textTheme.headlineMedium,
                            )),
                            const SizedBox(height: 30),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(localeMsg.selectOS),
                                const SizedBox(width: 20),
                                SizedBox(
                                  height: 35,
                                  width: 165,
                                  child: DropdownButtonFormField<CliOS>(
                                    decoration: GetFormInputDecoration(
                                      false,
                                      null,
                                      icon: Icons.desktop_windows,
                                    ),
                                    value: _selectedOS,
                                    items: CliOS.values
                                        .map<DropdownMenuItem<CliOS>>(
                                            (CliOS value) {
                                      return DropdownMenuItem<CliOS>(
                                        value: value,
                                        child: Text(
                                          value == CliOS.macOS
                                              ? value.name
                                              : value.name.capitalize(),
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      );
                                    }).toList(),
                                    onChanged: (CliOS? value) {
                                      setState(() {
                                        _selectedOS = value!;
                                      });
                                    },
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
                                        submitCreateOpenDcim(localeMsg),
                                    label: Text(localeMsg.download),
                                    icon: _isLoading
                                        ? Container(
                                            width: 24,
                                            height: 24,
                                            padding: const EdgeInsets.all(2.0),
                                            child:
                                                const CircularProgressIndicator(
                                              color: Colors.white,
                                              strokeWidth: 3,
                                            ),
                                          )
                                        : const Icon(Icons.download, size: 16))
                              ],
                            )
                          ],
                        ),
                      ))),
        ),
      ),
    );
  }

  submitCreateOpenDcim(AppLocalizations localeMsg) async {
    const urlPath =
        'https://github.com/ditrit/OGrEE-Core/releases/latest/download/';
    String cliName = "OGrEE_CLI";
    switch (_selectedOS) {
      case CliOS.windows:
        cliName = "${cliName}_Win.exe";
        break;
      case CliOS.linux:
        cliName = "${cliName}_Linux";
        break;
      case CliOS.macOS:
        cliName = "${cliName}_OSX";
        break;
    }
    if (kIsWeb) {
      html.AnchorElement(href: urlPath + cliName)
        ..setAttribute("download", cliName)
        ..click();
    } else {
      // Save to local filesystem
      setState(() {
        _isLoading = true;
      });
      final messenger = ScaffoldMessenger.of(context);
      final navigator = Navigator.of(context);
      final response = await http.get(Uri.parse(urlPath + cliName));
      navigator.pop();
      if (response.statusCode >= 200 && response.statusCode < 300) {
        var path = (await getApplicationDocumentsDirectory()).path;
        var fileName = '$path/$cliName';
        var file = File(fileName);
        for (var i = 1; await file.exists(); i++) {
          fileName = '$path/$cliName ($i)';
          file = File(fileName);
        }
        file.writeAsBytes(response.bodyBytes, flush: true).then((value) =>
            showSnackBar(ScaffoldMessenger.of(context),
                "${localeMsg.fileSavedTo} $fileName",
                copyText: fileName));
      } else {
        showSnackBar(messenger, localeMsg.unableDownload);
      }
      setState(() {
        _isLoading = false;
      });
    }
  }
}
