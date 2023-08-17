import 'package:flutter/material.dart';

import '../main.dart';

class LanguageToggle extends StatefulWidget {
  @override
  State<LanguageToggle> createState() => _LanguageToggleState();
}

class _LanguageToggleState extends State<LanguageToggle> {
  // languages = [FR, EN]
  var _selectedLanguage = [true, false];

  @override
  Widget build(BuildContext context) {
    if (MyApp.of(context) != null && MyApp.of(context)!.getLocale() == 'en') {
      _selectedLanguage = [false, true];
    }
    return Container(
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints(maxWidth: 60, maxHeight: 30),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.all(Radius.circular(8.0)),
      ),
      child: ToggleButtons(
        onPressed: (int index) {
          setState(() {
            for (int i = 0; i < _selectedLanguage.length; i++) {
              _selectedLanguage[i] = i == index;
            }
            if (index == 1) {
              MyApp.of(context)!.setLocale(const Locale('en'));
            } else {
              MyApp.of(context)!.setLocale(const Locale('fr', 'FR'));
            }
          });
        },
        selectedColor: Colors.white,
        fillColor: Colors.blue,
        borderRadius: const BorderRadius.all(Radius.circular(8)),
        borderWidth: 0,
        constraints: const BoxConstraints(minWidth: 30, minHeight: 30),
        isSelected: _selectedLanguage,
        children: const <Widget>[Text("FR"), Text("EN")],
      ),
    );
  }
}
