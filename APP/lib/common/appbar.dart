import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/pages/login_page.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/tenant_page.dart';
import 'package:ogree_app/widgets/common/language_toggle.dart';
import 'package:ogree_app/widgets/login/change_password_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/create_server_popup.dart';
import 'package:ogree_app/widgets/tools/download_tool_popup.dart';

AppBar myAppBar(context, userEmail, {isTenantMode = false}) {
  final localeMsg = AppLocalizations.of(context)!;
  Future logout() => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => const LoginPage(),
        ),
      );

  final List<PopupMenuEntry<String>> entries = <PopupMenuEntry<String>>[
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
          child: Text(backendType == BackendType.kubernetes
              ? localeMsg.addKube
              : localeMsg.addServer,),
        ),);
  } else {
    entries.insert(
        0,
        PopupMenuItem(
          value: Tools.unity.name,
          child: Text(localeMsg.downloadUnity),
        ),);
    entries.insert(
        0,
        PopupMenuItem(
          value: Tools.cli.name,
          child: Text(localeMsg.downloadCli),
        ),);
    if (isTenantAdmin) {
      entries.insert(
          0,
          PopupMenuItem(
            value: "tenant",
            child: Text(localeMsg.tenantParameters),
          ),);
    }
  }

  final bool isSmallDisplay = MediaQuery.of(context).size.width < 600;
  return AppBar(
    backgroundColor: Colors.grey.shade900,
    leadingWidth: 160,
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
                  color: Colors.white,),
            ),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => ProjectsPage(
                    userEmail: isTenantMode ? "admin" : userEmail,
                    isTenantMode: isTenantMode,),
              ),
            ),
          ),
          Badge(
            isLabelVisible: isTenantMode,
            label: const Text("ADMIN"),
          ),
        ],
      ),
    ),
    actions: [
      if (isSmallDisplay) Container() else Padding(
              padding: const EdgeInsets.only(right: 20),
              child: Row(
                children: [
                  if (backendType == BackendType.kubernetes) Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: Container(
                            decoration: BoxDecoration(
                              borderRadius:
                                  const BorderRadius.all(Radius.circular(8)),
                              border: Border.all(color: Colors.white),
                            ),
                            child: Badge(
                              backgroundColor: Colors.grey.shade900,
                              label: const Text("KUBE"),
                            ),
                          ),
                        ) else Container(),
                  Text(isTenantMode ? apiUrl : tenantName,
                      style: const TextStyle(color: Colors.white),),
                ],
              ),
            ),
      const Padding(
        padding: EdgeInsets.symmetric(vertical: 15),
        child: LanguageToggle(),
      ),
      const SizedBox(width: 17),
      PopupMenuButton<String>(
          onSelected: (value) {
            if (value == "logout") {
              logout();
            } else if (value == "new") {
              showCustomPopup(
                  context, CreateServerPopup(parentCallback: () {}),);
            } else if (value == "tenant") {
              Navigator.of(context).push(MaterialPageRoute(
                builder: (context) => const TenantPage(userEmail: "admin"),
              ),);
            } else if (value == Tools.unity.name) {
              showCustomPopup(context, const DownloadToolPopup(tool: Tools.unity),
                  isDismissible: true,);
            } else if (value == Tools.cli.name) {
              showCustomPopup(context, const DownloadToolPopup(tool: Tools.cli),
                  isDismissible: true,);
            } else {
              showCustomPopup(context, const ChangePasswordPopup());
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
              if (isSmallDisplay) Tooltip(
                      message: isTenantMode
                          ? (backendType == BackendType.kubernetes
                              ? "(KUBE) $apiUrl"
                              : apiUrl)
                          : tenantName,
                      triggerMode: TooltipTriggerMode.tap,
                      child: const Icon(
                        Icons.info_outline_rounded,
                        color: Colors.white,
                      ),) else Text(
                      isTenantMode ? "admin" : userEmail,
                      style: const TextStyle(color: Colors.white),
                    ),
            ],
          ),),
      const SizedBox(width: 40),
    ],
  );
}
