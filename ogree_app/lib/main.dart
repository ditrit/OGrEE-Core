import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/login_page.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();

  static _MyAppState? of(BuildContext context) =>
      context.findAncestorStateOfType<_MyAppState>();
}

class _MyAppState extends State<MyApp> {
  // App language control
  Locale _locale = const Locale('fr', 'FR');
  void setLocale(Locale value) {
    setState(() {
      _locale = value;
    });
  }

  String getLocale() => _locale.languageCode;

  // This widget is the root of our application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'OGrEE App',
      locale: _locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      theme: ThemeData(
          // useMaterial3: true,
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
