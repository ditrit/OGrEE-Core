import 'package:flag/flag.dart';
import 'package:flutter/material.dart';

import 'package:ogree_app/main.dart';

class LanguageOption {
  Flag flag;
  String name;
  Locale locale;

  LanguageOption(
      {required this.flag, required this.name, required this.locale,});
}

class LanguageToggle extends StatefulWidget {
  const LanguageToggle({super.key});

  @override
  State<LanguageToggle> createState() => _LanguageToggleState();
}

class _LanguageToggleState extends State<LanguageToggle> {
  Flag getFlag(String country) => Flag.fromString(
        country,
        height: 20,
        width: 20,
        fit: BoxFit.fill,
        flagSize: FlagSize.size_1x1,
        borderRadius: 10,
      );
  DecoratedBox flagWrapper(Flag flag) => DecoratedBox(
        decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(10.0),
            border: Border.all(
              color: Colors.white,
            ),),
        child: Padding(padding: const EdgeInsets.all(0.5), child: flag),
      );
  late List<LanguageOption> languages;
  late LanguageOption _selectedLanguage;

  @override
  void initState() {
    super.initState();
    languages = [
      LanguageOption(
          flag: getFlag("FR"),
          name: 'Français',
          locale: const Locale('fr', 'FR'),),
      LanguageOption(
          flag: getFlag("GB"), name: 'English', locale: const Locale('en'),),
      LanguageOption(
          flag: getFlag("ES"), name: 'Español', locale: const Locale('es'),),
      LanguageOption(
          flag: getFlag("BR"), name: 'Português', locale: const Locale('pt'),),
    ];
    _selectedLanguage = languages.first;
  }

  @override
  Widget build(BuildContext context) {
    if (MyApp.of(context) != null &&
        MyApp.of(context)!.getLocale() != _selectedLanguage.locale) {
      _selectedLanguage = languages.firstWhere(
          (language) => language.locale == MyApp.of(context)!.getLocale(),);
    }
    return PopupMenuButton(
      child: flagWrapper(_selectedLanguage.flag),
      onSelected: (language) => setState(() {
        _selectedLanguage = language;
        MyApp.of(context)!.setLocale(_selectedLanguage.locale);
      }),
      itemBuilder: (context) {
        return languages.map((language) {
          return PopupMenuItem(
            value: language,
            child: Row(
              children: [
                flagWrapper(language.flag),
                Padding(
                  padding: const EdgeInsets.only(left: 8.0),
                  child: Text(language.name),
                ),
              ],
            ),
          );
        }).toList();
      },
    );
  }
}
