import 'dart:convert';
import 'dart:io';
import 'package:csv/csv.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:path_provider/path_provider.dart';
import 'package:universal_html/html.dart' as html;
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

saveCSV(String desiredFileName, List<List<String>> rows,
    BuildContext context) async {
  // Prepare the file
  final String csv = const ListToCsvConverter().convert(rows);
  final bytes = utf8.encode(csv);
  if (kIsWeb) {
    // If web, use html to download csv
    html.AnchorElement(
      href: 'data:application/octet-stream;base64,${base64Encode(bytes)}',
    )
      ..setAttribute("download", "$desiredFileName.csv")
      ..click();
  } else {
    // Save to local filesystem
    final localeMsg = AppLocalizations.of(context)!;
    final messenger = ScaffoldMessenger.of(context);
    final path = (await getApplicationDocumentsDirectory()).path;
    var fileName = '$path/$desiredFileName.csv';
    var file = File(fileName);
    for (var i = 1; await file.exists(); i++) {
      fileName = '$path/$desiredFileName ($i).csv';
      file = File(fileName);
    }
    file.writeAsBytes(bytes, flush: true).then(
          (value) => showSnackBar(
            messenger,
            "${localeMsg.fileSavedTo} $fileName",
          ),
        );
  }
}
