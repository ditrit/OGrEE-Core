import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

AppBar myAppBar() {
  return AppBar(
    backgroundColor: Colors.grey.shade900,
    leadingWidth: 150,
    leading: Center(
        child: Text(
      'OGrEE',
      style: GoogleFonts.inter(
        fontSize: 21,
        fontWeight: FontWeight.w700,
      ),
    )),
    actions: const [
      Icon(Icons.account_circle),
      SizedBox(
        width: 15,
      ),
      Center(child: Text('DOE John')),
      SizedBox(
        width: 50,
      )
    ],
  );
}
