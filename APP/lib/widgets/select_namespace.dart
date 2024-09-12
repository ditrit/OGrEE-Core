import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/pages/select_page.dart';

class SelectNamespace extends StatefulWidget {
  const SelectNamespace({super.key});
  @override
  State<SelectNamespace> createState() => _SelectNamespaceState();
}

class _SelectNamespaceState extends State<SelectNamespace> {
  Map<String, String> namespaces = {
    Namespace.Physical.name: "site.building.room",
    Namespace.Organisational.name: "domains",
    Namespace.Logical.name: "templates&groups",
  };
  Namespace _selection = Namespace.Physical;

  @override
  void initState() {
    //SelectPage.of(context)!.selectedNamespace = _selection;
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
        SizedBox(
          width: MediaQuery.of(context).size.width > 50
              ? MediaQuery.of(context).size.width - 50
              : MediaQuery.of(context).size.width,
          child: Card(
              child: Wrap(
            alignment: WrapAlignment.spaceEvenly,
            crossAxisAlignment: WrapCrossAlignment.center,
            children:
                namespaces.keys.map((label) => nameSpaceButton(label)).toList(),
          ),),
        ),
      ],
    );
  }

  Widget nameSpaceButton(String label) {
    final isBigScreen = MediaQuery.of(context).size.width > 800;
    final Namespace thisNamespace =
        Namespace.values.firstWhere((e) => e.toString() == 'Namespace.$label');
    return Container(
      margin: const EdgeInsets.only(top: 30, bottom: 30),
      width: isBigScreen ? 250 : 200,
      height: isBigScreen ? 100 : 70,
      child: OutlinedButton(
        onPressed: () => setState(() {
          _selection = thisNamespace;
          SelectPage.of(context)!.selectedNamespace = _selection;
        }),
        style: _selection == thisNamespace
            ? OutlinedButton.styleFrom(
                side: const BorderSide(width: 3.0, color: Colors.blue),
              )
            : OutlinedButton.styleFrom(
                side: BorderSide(color: Colors.blue.shade800),
              ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              label,
              style: GoogleFonts.inter(
                fontSize: 17,
                color: _selection == thisNamespace ? Colors.blue : null,
              ),
            ),
            Text(
              isBigScreen ? '\n${namespaces[label]}' : namespaces[label]!,
              style: GoogleFonts.inter(
                color: _selection == thisNamespace ? Colors.blue : null,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
