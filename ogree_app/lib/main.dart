import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/login_page.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of our application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'OGrEE App',
      theme: ThemeData(
          fontFamily: GoogleFonts.inter().fontFamily,
          textTheme: TextTheme(
            headlineLarge: GoogleFonts.inter(
              fontSize: 23,
              color: Colors.black,
              fontWeight: FontWeight.w700,
            ),
            headlineMedium: GoogleFonts.inter(
              fontSize: 17,
              color: Colors.black,
            ),
          )),
      home: const LoginPage(),
    );
  }
}
