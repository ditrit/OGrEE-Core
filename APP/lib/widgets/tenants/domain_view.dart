import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/settings_view.dart';
import 'package:ogree_app/widgets/select_objects/tree_view/custom_tree_view.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';

class DomainView extends StatefulWidget {
  const DomainView({super.key});

  @override
  State<DomainView> createState() => _DomainViewState();
}

class _DomainViewState extends State<DomainView> {
  final TreeAppController appController = TreeAppController();
  bool _reloadDomains = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Stack(children: [
      TreeAppControllerScope(
        controller: appController,
        child: FutureBuilder<void>(
          future: appController.init({},
              argNamespace: Namespace.Organisational,
              reload: _reloadDomains,
              isTenantMode: true,),
          builder: (_, __) {
            if (_reloadDomains) {
              _reloadDomains = false;
            }
            if (appController.isInitialized) {
              if (appController.treeController.roots.isEmpty) {
                return Column(
                  children: [
                    Icon(
                      Icons.warning_rounded,
                      size: 50,
                      color: Colors.grey.shade600,
                    ),
                    Padding(
                      padding: const EdgeInsets.only(top: 16),
                      child: Text("${localeMsg.noObjectsFound} :("),
                    ),
                  ],
                );
              }
              return Stack(children: [
                const CustomTreeView(isTenantMode: true),
                if (isSmallDisplay) Container() else const Align(
                        alignment: Alignment.topRight,
                        child: Padding(
                          padding: EdgeInsets.only(right: 16),
                          child: SizedBox(
                              width: 320,
                              height: 116,
                              child: Card(
                                  child: SettingsView(
                                isTenantMode: true,
                                namespace: Namespace.Organisational,
                              ),),),
                        ),
                      ),
              ],);
            }
            return const Center(child: CircularProgressIndicator());
          },
        ),
      ),
      Align(
        alignment: Alignment.bottomRight,
        child: Padding(
          padding: const EdgeInsets.only(bottom: 20, right: 20),
          child: ElevatedButton.icon(
            onPressed: () =>
                showCustomPopup(context, DomainPopup(parentCallback: () {
              setState(() {
                _reloadDomains = true;
              });
            },),),
            icon: const Icon(Icons.add),
            label: Text("${localeMsg.create} ${localeMsg.domain}"),
          ),
        ),
      ),
    ],);
  }
}
