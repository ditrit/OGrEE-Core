import 'package:flutter/material.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'settings_view/settings_view.dart';
import 'tree_view/custom_tree_view.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class SelectObjects extends StatefulWidget {
  final String dateRange;
  final String namespace;
  final bool load;

  const SelectObjects(
      {super.key,
      required this.dateRange,
      required this.namespace,
      required this.load});
  @override
  State<SelectObjects> createState() => _SelectObjectsState();
}

class _SelectObjectsState extends State<SelectObjects> {
  late final AppController appController = AppController();

  @override
  Widget build(BuildContext context) {
    return AppControllerScope(
      controller: appController,
      child: FutureBuilder<void>(
        future: widget.load
            ? appController.init(
                widget.namespace == Namespace.Test.name
                    ? {}
                    : SelectPage.of(context)!.selectedObjects,
                dateRange: widget.dateRange,
                isTest: widget.namespace == Namespace.Test.name,
                reload: widget.load,
                onlyDomain: widget.namespace == Namespace.Organisational.name)
            : null,
        builder: (_, __) {
          print(widget.load);
          if (appController.isInitialized && widget.load) {
            return _Unfocus(
              child: Card(
                margin: const EdgeInsets.all(0.1),
                child: appController.rootNode.children.isEmpty
                    ? Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.warning_rounded,
                            size: 50,
                            color: Colors.grey.shade600,
                          ),
                          Padding(
                            padding: const EdgeInsets.only(top: 16),
                            child: Text(
                                AppLocalizations.of(context)!.noObjectsFound +
                                    " :("),
                          ),
                        ],
                      )
                    : _ResponsiveBody(
                        onlyDomain:
                            widget.namespace == Namespace.Organisational.name,
                        controller: appController),
              ),
            );
          }
          return const Center(child: CircularProgressIndicator());
        },
      ),
    );
  }
}

class _Unfocus extends StatelessWidget {
  const _Unfocus({Key? key, required this.child}) : super(key: key);

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: FocusScope.of(context).unfocus,
      child: child,
    );
  }
}

class _ResponsiveBody extends StatelessWidget {
  final bool onlyDomain;
  final AppController controller;
  const _ResponsiveBody(
      {Key? key, required this.controller, this.onlyDomain = false})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    // print("BUILD RespBody " + MediaQuery.of(context).size.width.toString());
    if (MediaQuery.of(context).size.width < 600 &&
        MediaQuery.of(context).size.width != 0) {
      return Stack(
        children: [
          const CustomTreeView(isTenantMode: false),
          Align(
            alignment: Alignment.bottomRight,
            child: Padding(
              padding: const EdgeInsets.only(bottom: 20, right: 20),
              child: ElevatedButton.icon(
                onPressed: () => showCustomPopup(
                    context,
                    SettingsViewPopup(
                      controller: controller,
                    ),
                    isDismissible: true),
                icon: const Icon(Icons.filter_alt_outlined),
                label: Text(AppLocalizations.of(context)!.filters),
              ),
            ),
          ),
        ],
      );
    }
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        children: [
          const Flexible(flex: 2, child: CustomTreeView(isTenantMode: false)),
          const VerticalDivider(
            width: 1,
            thickness: 1,
            color: Colors.black26,
          ),
          Expanded(
              child: SettingsView(isTenantMode: false, noFilters: onlyDomain)),
        ],
      ),
    );
  }
}

class SettingsViewPopup extends StatelessWidget {
  final AppController controller;

  SettingsViewPopup({super.key, required this.controller});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        height: 500,
        child: AppControllerScope(
            controller: controller,
            child: Container(
                width: 500,
                constraints: const BoxConstraints(maxHeight: 625),
                margin: const EdgeInsets.symmetric(horizontal: 20),
                decoration: PopupDecoration,
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 20, 30, 15),
                  child: Material(
                      color: Colors.white,
                      child: ListView(
                        padding: EdgeInsets.zero,
                        shrinkWrap: true,
                        children: [
                          const SizedBox(
                            height: 420,
                            child: SettingsView(isTenantMode: false),
                          ),
                          const SizedBox(height: 10),
                          TextButton.icon(
                            style: OutlinedButton.styleFrom(
                                foregroundColor: Colors.blue.shade900),
                            onPressed: () => Navigator.pop(context),
                            label: Text(AppLocalizations.of(context)!.close),
                            icon: const Icon(
                              Icons.cancel_outlined,
                              size: 16,
                            ),
                          ),
                        ],
                      )),
                ))),
      ),
    );
  }
}
