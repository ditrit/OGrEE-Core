import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

const Map<String, String> namespaces = {
  "Physical": "site.building.room",
  "Logical": "app.cluster.lorem",
  "Organisational": "service.lorem.ipsum"
};

class SelectNamespace extends StatefulWidget {
  const SelectNamespace({super.key});
  @override
  State<SelectNamespace> createState() => _SelectNamespaceState();
}

class _SelectNamespaceState extends State<SelectNamespace> {
  String _selection = namespaces.keys.first;

  @override
  void initState() {
    SelectPage.of(context)!.selectedNamespace = _selection;
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Column(
      children: [
        Text(
          localeMsg.whatNamespace,
          style: Theme.of(context).textTheme.headlineLarge,
        ),
        const SizedBox(height: 25),
        Card(
            child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children:
              namespaces.keys.map((label) => NameSpaceButton(label)).toList(),
        )),
      ],
    );
  }

  Widget NameSpaceButton(String label) {
    return Container(
      margin: const EdgeInsets.only(top: 30, bottom: 30),
      width: 250,
      height: 100.0,
      child: OutlinedButton(
        onPressed: () => setState(() {
          _selection = label;
          SelectPage.of(context)!.selectedNamespace = _selection;
        }),
        style: _selection == label
            ? OutlinedButton.styleFrom(
                side: const BorderSide(width: 3.0, color: Colors.blue),
              )
            : OutlinedButton.styleFrom(
                side: const BorderSide(width: 0.5, color: Colors.grey),
              ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              label,
              style: GoogleFonts.inter(
                fontSize: 17,
                color: _selection == label ? Colors.blue : Colors.black,
              ),
            ),
            Text(
              '\n${namespaces[label]}',
              style: GoogleFonts.inter(
                color: _selection == label ? Colors.blue : Colors.black,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
