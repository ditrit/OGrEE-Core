import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

Future<Map<String, String>> getFrenchMessages() async {
  final String response = await rootBundle.loadString('lib/l10n/app_fr.arb');
  Map<String, dynamic> data = json.decode(response);
  Map<String, String> resp = {};
  for (String key in data.keys) {
    resp[key] = data[key].toString();
  }
  return resp;
}

class LocalizationsInjApp extends StatelessWidget {
  final Widget child;
  const LocalizationsInjApp({Key? key, required this.child}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      locale: const Locale('fr', 'FR'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: child,
    );
  }
}

class LocalizationsInj extends StatelessWidget {
  final Widget child;
  const LocalizationsInj({Key? key, required this.child}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Localizations(
      locale: const Locale('fr', 'FR'),
      delegates: AppLocalizations.localizationsDelegates,
      child: child,
    );
  }
}
