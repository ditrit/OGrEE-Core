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

enum PopupMenuEntries {
  passwordChange,
  logout,
  createNewServer,
  tenantParams,
  downloadUnity,
  downloadCli
}

AppBar myAppBar(BuildContext context, String userEmail,
    {bool isTenantMode = false}) {
  final localeMsg = AppLocalizations.of(context)!;
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
                color: Colors.white,
              ),
            ),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => ProjectsPage(
                  userEmail: isTenantMode ? "admin" : userEmail,
                  isTenantMode: isTenantMode,
                ),
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
      getInfoBadge(isTenantMode, isSmallDisplay),
      const Padding(
        padding: EdgeInsets.symmetric(vertical: 15),
        child: LanguageToggle(),
      ),
      const SizedBox(width: 17),
      getPopupMenuButton(
          isTenantMode, isSmallDisplay, userEmail, localeMsg, context),
      const SizedBox(width: 40),
    ],
  );
}

// KUBE + API URL + TenantName
Widget getInfoBadge(bool isTenantMode, bool isSmallDisplay) {
  if (isSmallDisplay) {
    return Container();
  } else {
    return Padding(
      padding: const EdgeInsets.only(right: 20),
      child: Row(
        children: [
          if (backendType == BackendType.kubernetes)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Container(
                decoration: BoxDecoration(
                  borderRadius: const BorderRadius.all(Radius.circular(8)),
                  border: Border.all(color: Colors.white),
                ),
                child: Badge(
                  backgroundColor: Colors.grey.shade900,
                  label: const Text("KUBE"),
                ),
              ),
            )
          else
            Container(),
          Text(
            isTenantMode ? apiUrl : tenantName,
            style: const TextStyle(color: Colors.white),
          ),
        ],
      ),
    );
  }
}

// POPUP MENU
PopupMenuButton<PopupMenuEntries> getPopupMenuButton(
    bool isTenantMode,
    bool isSmallDisplay,
    String userEmail,
    AppLocalizations localeMsg,
    BuildContext context) {
  return PopupMenuButton<PopupMenuEntries>(
    onSelected: (value) => onMenuEntrySelected(value, context),
    itemBuilder: (_) => getPopupMenuEntries(isTenantMode, localeMsg),
    child: Row(
      children: [
        const Icon(
          Icons.account_circle,
          color: Colors.white,
        ),
        const SizedBox(width: 10),
        if (isSmallDisplay)
          Tooltip(
            message: isTenantMode
                ? (backendType == BackendType.kubernetes
                    ? "(KUBE) $apiUrl"
                    : apiUrl)
                : tenantName,
            triggerMode: TooltipTriggerMode.tap,
            child: const Icon(
              Icons.info_outline_rounded,
              color: Colors.white,
            ),
          )
        else
          Text(
            isTenantMode ? "admin" : userEmail,
            style: const TextStyle(color: Colors.white),
          ),
      ],
    ),
  );
}

List<PopupMenuEntry<PopupMenuEntries>> getPopupMenuEntries(
    bool isTenantMode, AppLocalizations localeMsg) {
  final List<PopupMenuEntry<PopupMenuEntries>> entries =
      <PopupMenuEntry<PopupMenuEntries>>[
    PopupMenuItem(
      value: PopupMenuEntries.passwordChange,
      child: Text(localeMsg.changePassword),
    ),
    const PopupMenuItem(
      value: PopupMenuEntries.logout,
      child: Text("Logout"),
    ),
  ];
  if (isTenantMode) {
    entries.insert(
      0,
      PopupMenuItem(
        value: PopupMenuEntries.createNewServer,
        child: Text(
          backendType == BackendType.kubernetes
              ? localeMsg.addKube
              : localeMsg.addServer,
        ),
      ),
    );
  } else {
    entries.insert(
      0,
      PopupMenuItem(
        value: PopupMenuEntries.downloadUnity,
        child: Text(localeMsg.downloadUnity),
      ),
    );
    entries.insert(
      0,
      PopupMenuItem(
        value: PopupMenuEntries.downloadCli,
        child: Text(localeMsg.downloadCli),
      ),
    );
    if (isTenantAdmin) {
      entries.insert(
        0,
        PopupMenuItem(
          value: PopupMenuEntries.tenantParams,
          child: Text(localeMsg.tenantParameters),
        ),
      );
    }
  }
  return entries;
}

onMenuEntrySelected(PopupMenuEntries selectedEntry, BuildContext context) {
  switch (selectedEntry) {
    case PopupMenuEntries.logout:
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => const LoginPage(),
        ),
      );
      break;
    case PopupMenuEntries.tenantParams:
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => const TenantPage(userEmail: "admin"),
        ),
      );
      break;
    case PopupMenuEntries.downloadUnity:
      showCustomPopup(
        context,
        const DownloadToolPopup(tool: Tools.unity),
        isDismissible: true,
      );
      break;
    case PopupMenuEntries.downloadCli:
      showCustomPopup(
        context,
        const DownloadToolPopup(tool: Tools.cli),
        isDismissible: true,
      );
      break;
    case PopupMenuEntries.passwordChange:
      showCustomPopup(context, const ChangePasswordPopup());
      break;
    case PopupMenuEntries.createNewServer:
      showCustomPopup(
        context,
        CreateServerPopup(parentCallback: () {}),
      );
      break;
  }
}
