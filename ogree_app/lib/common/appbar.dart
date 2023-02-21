import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/projects_page.dart';

AppBar myAppBar(context) {
  return AppBar(
    backgroundColor: Colors.grey.shade900,
    leadingWidth: 150,
    leading: Center(
        child: TextButton(
      child: Text(
        'OGrEE',
        style: GoogleFonts.inter(
            fontSize: 21, fontWeight: FontWeight.w700, color: Colors.white),
      ),
      onPressed: () => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => ProjectsPage(),
        ),
      ),
    )),
    actions: const [
      Icon(Icons.account_circle),
      SizedBox(
        width: 15,
      ),
      Center(child: Text('Admin')),
      SizedBox(
        width: 50,
      )
    ],
  );
}
