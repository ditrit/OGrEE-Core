import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

void showSnackBar(
  ScaffoldMessengerState messenger,
  String message, {
  Duration duration = const Duration(seconds: 6),
  bool isError = false,
  bool isSuccess = false,
  String copyText = "",
}) {
  var color = Colors.blueGrey.shade900;
  if (isError) color = Colors.red.shade900;
  if (isSuccess) color = Colors.green;
  messenger
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: color,
        content: Text(message),
        duration: duration,
        action: copyText == ""
            ? null
            : SnackBarAction(
                label: "COPY",
                onPressed: () =>
                    Clipboard.setData(ClipboardData(text: copyText))),
        showCloseIcon: duration.inSeconds > 5,
      ),
    );
}
