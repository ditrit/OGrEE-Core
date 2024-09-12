import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/login_page.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
    await InAppWebViewController.setWebContentsDebuggingEnabled(kDebugMode);
  }

  await dotenv.load(fileName: "assets/custom/.env");
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => MyAppState();

  static MyAppState? of(BuildContext context) =>
      context.findAncestorStateOfType<MyAppState>();
}

class MyAppState extends State<MyApp> {
  // App language control
  Locale _locale = const Locale('fr', 'FR');
  void setLocale(Locale value) {
    setState(() {
      _locale = value;
    });
  }

  Locale getLocale() => _locale;

  // This widget is the root of our application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      // debugShowCheckedModeBanner: false,
      title: 'OGrEE App',
      locale: _locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      theme: ThemeData(
        useMaterial3: true,
        colorSchemeSeed: Colors.blue,
        fontFamily: GoogleFonts.inter().fontFamily,
        scrollbarTheme: ScrollbarThemeData(
          thumbVisibility: WidgetStateProperty.all<bool>(true),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: Colors.blue.shade600,
            foregroundColor: Colors.white,
          ),
        ),
        cardTheme: const CardTheme(
          elevation: 3,
          surfaceTintColor: Colors.white,
          color: Colors.white,
        ),
        textTheme: TextTheme(
          headlineLarge: GoogleFonts.inter(
            fontSize: 22,
            color: Colors.black,
            fontWeight: FontWeight.w700,
          ),
          headlineMedium: GoogleFonts.inter(
            fontSize: 20,
            color: Colors.black,
            fontWeight: FontWeight.w400,
          ),
          headlineSmall: GoogleFonts.inter(
            fontSize: 17,
            color: Colors.black,
          ),
        ),
      ),
      home: const LoginPage(),
      onGenerateRoute: RouteGenerator.generateRoute,
    );
  }
}

class RouteGenerator {
  static Route<dynamic> generateRoute(RouteSettings settings) {
    Map? queryParameters;
    if (settings.name != null) {
      final uriData = Uri.parse(settings.name!);
      queryParameters = uriData.queryParameters;
    }
    return MaterialPageRoute(
      builder: (context) {
        return LoginPage(
          isPasswordReset: true,
          resetToken: queryParameters!["token"].toString(),
        );
      },
      settings: settings,
    );
  }
}
