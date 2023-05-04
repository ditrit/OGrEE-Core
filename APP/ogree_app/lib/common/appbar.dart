import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/language_toggle.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import '../widgets/tenants/popups/create_server_popup.dart';

AppBar myAppBar(context, userEmail, {isTenantMode = false}) {
  logout() => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => const LoginPage(),
        ),
      );
  return AppBar(
    backgroundColor: Colors.grey.shade900,
    leadingWidth: 150,
    leading: Center(
        child: TextButton(
      child: Text(
        'OGrEE',
        style: TextStyle(
            fontSize: 21, fontWeight: FontWeight.w700, color: Colors.white),
      ),
      onPressed: () => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => ProjectsPage(
              userEmail: isTenantMode ? "admin" : userEmail,
              isTenantMode: isTenantMode),
        ),
      ),
    )),
    actions: [
      isTenantMode
          ? Padding(
              padding: const EdgeInsets.only(right: 20),
              child: Text(apiUrl, style: TextStyle(color: Colors.white)),
            )
          : Container(),
      Padding(
        padding: const EdgeInsets.symmetric(vertical: 15),
        child: LanguageToggle(),
      ),
      const SizedBox(width: 20),
      PopupMenuButton<String>(
          onSelected: (value) {
            if (value == "logout") {
              logout();
            } else {
              showCustomPopup(
                  context, CreateServerPopup(parentCallback: () {}));
            }
          },
          itemBuilder: (_) => <PopupMenuEntry<String>>[
                PopupMenuItem(
                  value: "logout",
                  child: Text("Logout"),
                ),
                PopupMenuItem(
                  value: "new",
                  child: Text(AppLocalizations.of(context)!.addServer),
                ),
              ],
          child: Row(
            children: [
              const Icon(
                Icons.account_circle,
                color: Colors.white,
              ),
              const SizedBox(width: 10),
              Text(
                isTenantMode ? "admin" : userEmail,
                style: TextStyle(color: Colors.white),
              ),
            ],
          )),
      const SizedBox(width: 40)
    ],
  );
}
