import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/language_toggle.dart';

AppBar myAppBar(context, userEmail) {
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
          builder: (context) => ProjectsPage(userEmail: userEmail),
        ),
      ),
    )),
    actions: [
      Padding(
        padding: const EdgeInsets.symmetric(vertical: 15),
        child: LanguageToggle(),
      ),
      const SizedBox(width: 20),
      const Icon(Icons.account_circle),
      const SizedBox(width: 10),
      Center(child: Text(userEmail)),
      const SizedBox(width: 40)
    ],
  );
}
