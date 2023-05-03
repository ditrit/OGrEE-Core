import 'package:flutter/material.dart';

void showSnackBar(
  BuildContext context,
  String message, {
  Duration duration = const Duration(seconds: 3),
  bool isError = false,
  bool isSuccess = false,
}) {
  var color = Colors.blueGrey.shade900;
  if (isError) color = Colors.red.shade900;
  if (isSuccess) color = Colors.green;
  ScaffoldMessenger.of(context)
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: color,
        content: Text(message),
        duration: duration,
        showCloseIcon: duration.inSeconds > 5,
      ),
    );
}
