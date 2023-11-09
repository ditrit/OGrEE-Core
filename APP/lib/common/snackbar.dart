import 'package:flutter/material.dart';

void showSnackBar(
  ScaffoldMessengerState messenger,
  String message, {
  Duration duration = const Duration(seconds: 6),
  bool isError = false,
  bool isSuccess = false,
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
        showCloseIcon: duration.inSeconds > 5,
      ),
    );
}
