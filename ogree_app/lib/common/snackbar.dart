import 'package:flutter/material.dart';

void showSnackBar(
  BuildContext context,
  String message, {
  Duration duration = const Duration(seconds: 2),
  bool isError = false,
}) {
  ScaffoldMessenger.of(context)
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor:
            isError ? Colors.red.shade900 : Colors.blueGrey.shade900,
        content: Text(message),
        duration: duration,
      ),
    );
}
