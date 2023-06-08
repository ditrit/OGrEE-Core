import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/change_password_popup.dart';
import 'package:ogree_app/widgets/language_toggle.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import '../widgets/tenants/popups/create_server_popup.dart';

AppBar myAppBar(context, userEmail, {isTenantMode = false}) {
  logout() => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => const LoginPage(),
        ),
      );

  List<PopupMenuEntry<String>> entries = <PopupMenuEntry<String>>[
    PopupMenuItem(
      value: "change",
      child: Text(AppLocalizations.of(context)!.changePassword),
    ),
    const PopupMenuItem(
      value: "logout",
      child: Text("Logout"),
    ),
  ];
  if (isTenantMode) {
    entries.insert(
        0,
        PopupMenuItem(
          value: "new",
          child: Text(AppLocalizations.of(context)!.addServer),
        ));
  }

  return AppBar(
    backgroundColor: Colors.grey.shade900,
    leadingWidth: 150,
    leading: Padding(
      padding: const EdgeInsets.only(left: 20),
      child: Row(
        children: [
          TextButton(
            child: const Text(
              'OGrEE',
              style: TextStyle(
                  fontSize: 21,
                  fontWeight: FontWeight.w700,
                  color: Colors.white),
            ),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => ProjectsPage(
                    userEmail: isTenantMode ? "admin" : userEmail,
                    isTenantMode: isTenantMode),
              ),
            ),
          ),
          Badge(
            isLabelVisible: isTenantMode,
            label: Text("ADMIN"),
          )
        ],
      ),
    ),
    actions: [
      Padding(
        padding: const EdgeInsets.only(right: 20),
        child: Text(apiUrl, style: const TextStyle(color: Colors.white)),
      ),
      Padding(
        padding: const EdgeInsets.symmetric(vertical: 15),
        child: LanguageToggle(),
      ),
      const SizedBox(width: 20),
      PopupMenuButton<String>(
          onSelected: (value) {
            if (value == "logout") {
              logout();
            } else if (value == "new") {
              showCustomPopup(
                  context, CreateServerPopup(parentCallback: () {}));
            } else {
              showCustomPopup(context, ChangePasswordPopup());
            }
          },
          itemBuilder: (_) => entries,
          child: Row(
            children: [
              const Icon(
                Icons.account_circle,
                color: Colors.white,
              ),
              const SizedBox(width: 10),
              Text(
                isTenantMode ? "admin" : userEmail,
                style: const TextStyle(color: Colors.white),
              ),
            ],
          )),
      const SizedBox(width: 40)
    ],
  );
}
