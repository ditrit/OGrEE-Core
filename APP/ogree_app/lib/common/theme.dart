import 'package:flutter/material.dart';

GetFormInputDecoration(isSmallDisplay, String? labelText,
        {IconData? icon,
        Color? iconColor,
        String? prefixText,
        String? suffixText,
        String? hint}) =>
    InputDecoration(
      prefixIcon: isSmallDisplay
          ? null
          : Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12.0),
              child: Icon(
                icon,
                color: iconColor ?? Colors.grey.shade400,
                // color: Colors.blue.shade600,
              ),
            ),
      prefixText: prefixText,
      suffixText: suffixText,
      labelText: labelText,
      hintText: hint,
      labelStyle: const TextStyle(
        fontSize: 14.0,
      ),
      filled: true,
      fillColor: const Color.fromARGB(255, 248, 247, 247),
      contentPadding: const EdgeInsets.only(
        top: 3.0,
        bottom: 12.0,
        left: 20.0,
        right: 14.0,
      ),
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
