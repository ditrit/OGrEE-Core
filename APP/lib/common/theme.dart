// ignore_for_file: non_constant_identifier_names, constant_identifier_names

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/definitions.dart';

InputDecoration GetFormInputDecoration(
  isSmallDisplay,
  String? labelText, {
  IconData? icon,
  Color? iconColor,
  String? prefixText,
  String? suffixText,
  String? hint,
  EdgeInsets? contentPadding = const EdgeInsets.only(
    top: 3.0,
    bottom: 12.0,
    left: 20.0,
    right: 14.0,
  ),
  bool isEnabled = true,
  bool isCompact = false,
  Widget? iconWidget,
}) =>
    InputDecoration(
      prefixIcon: iconWidget ??
          (isSmallDisplay
              ? null
              : Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12.0),
                  child: Icon(
                    icon,
                    color: iconColor ?? Colors.grey.shade400,
                    // color: Colors.blue.shade600,
                  ),
                )),
      prefixText: prefixText,
      suffixText: suffixText,
      label: labelText == null || labelText.isEmpty
          ? null
          : (labelText[0] == starSymbol
              ? Row(
                  children: [
                    RichText(
                      text: TextSpan(
                        text: labelText.replaceFirst(starSymbol, ""),
                        style: const TextStyle(color: Colors.black),
                        children: const [
                          TextSpan(
                            text: starSymbol,
                            style: TextStyle(color: Colors.red),
                          ),
                        ],
                      ),
                    ),
                  ],
                )
              : Text(labelText)),
      hintText: hint,
      enabled: isEnabled,
      labelStyle: const TextStyle(
        fontSize: 14.0,
      ),
      errorStyle: isCompact ? const TextStyle(fontSize: 9, height: 0.25) : null,
      filled: true,
      fillColor: const Color.fromARGB(255, 248, 247, 247),
      contentPadding: contentPadding,
      border: UnderlineInputBorder(
        borderRadius: BorderRadius.circular(12.0),
        borderSide: BorderSide.none,
      ),
    );

const FormInputPadding = EdgeInsets.only(left: 2, right: 10, bottom: 8, top: 2);

final PopupDecoration = BoxDecoration(
  color: Colors.white,
  borderRadius: BorderRadius.circular(30),
  boxShadow: const [
    // Shadow for top-left corner
    BoxShadow(
      color: Colors.grey,
      offset: Offset(10, 10),
      blurRadius: 6,
      spreadRadius: 1,
    ),
    // Shadow for bottom-right corner
    BoxShadow(
      color: Colors.white12,
      offset: Offset(-10, -10),
      blurRadius: 5,
      spreadRadius: 1,
    ),
  ],
);

bool IsSmallDisplay(double width) => width < 550;

InputDecoration LoginInputDecoration({
  required String label,
  String? hint,
  bool isSmallDisplay = false,
}) =>
    InputDecoration(
      contentPadding: isSmallDisplay
          ? const EdgeInsets.symmetric(horizontal: 12, vertical: 16)
          : null,
      labelText: label,
      hintText: hint,
      labelStyle: GoogleFonts.inter(
        fontSize: 11,
        color: Colors.black,
      ),
      border: const OutlineInputBorder(
        borderSide: BorderSide(
          color: Colors.grey,
        ),
      ),
    );
