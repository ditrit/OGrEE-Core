import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

GetFormInputDecoration(isSmallDisplay, String? labelText,
        {IconData? icon,
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
        Widget? iconWidget}) =>
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
      labelText: labelText,
      hintText: hint,
      enabled: isEnabled,
      labelStyle: const TextStyle(
        fontSize: 14.0,
      ),
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
    ]);

IsSmallDisplay(width) => width < 550;

LoginInputDecoration(
        {required String label, String? hint, bool isSmallDisplay = false}) =>
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
          width: 1,
        ),
      ),
    );
