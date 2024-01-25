import 'package:flutter/material.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/object_popup.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';
import 'settings_view/settings_view.dart';
import 'tree_view/custom_tree_view.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class SelectObjects extends StatefulWidget {
  final String dateRange;
  final Namespace namespace;
  bool load;

  SelectObjects(
      {super.key,
      required this.dateRange,
      required this.namespace,
      required this.load});
  @override
  State<SelectObjects> createState() => _SelectObjectsState();
}

class _SelectObjectsState extends State<SelectObjects> {
  late final TreeAppController appController = TreeAppController();

  @override
  Widget build(BuildContext context) {
    return TreeAppControllerScope(
      controller: appController,
      child: FutureBuilder<void>(
        future: widget.load
            ? appController.init(
                widget.namespace == Namespace.Test
                    ? {}
                    : SelectPage.of(context)!.selectedObjects,
                dateRange: widget.dateRange,
                reload: widget.load,
                argNamespace: widget.namespace)
            : null,
        builder: (_, __) {
          print(widget.load);
          if (appController.isInitialized && widget.load) {
            return _Unfocus(
              child: Card(
                margin: const EdgeInsets.all(0.1),
                child: appController.treeController.roots.isEmpty
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
                                "${AppLocalizations.of(context)!.noObjectsFound} :("),
                          ),
                        ],
                      )
                    : _ResponsiveBody(
                        namespace: widget.namespace,
                        noFilters: widget.namespace != Namespace.Physical,
                        controller: appController,
                        callback: () => setState(() {
                              widget.load = true;
                            })),
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
  final Namespace namespace;
  final bool noFilters;
  final TreeAppController controller;
  final Function() callback;
  const _ResponsiveBody(
      {Key? key,
      required this.namespace,
      required this.controller,
      this.noFilters = false,
      required this.callback})
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
      child: Stack(
        children: [
          Row(
            children: [
              const Flexible(
                  flex: 2, child: CustomTreeView(isTenantMode: false)),
              const VerticalDivider(
                width: 1,
                thickness: 1,
                color: Colors.black26,
              ),
              Expanded(
                  child:
                      SettingsView(isTenantMode: false, noFilters: noFilters)),
            ],
          ),
          Padding(
            padding: const EdgeInsets.only(left: 6, bottom: 6),
            child: Align(
              alignment: Alignment.bottomLeft,
              child: SizedBox(
                height: 34,
                width: 34,
                child: IconButton(
                  padding: EdgeInsets.all(0.0),
                  iconSize: 24,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue.shade600,
                    foregroundColor: Colors.white,
                  ),
                  onPressed: () => showCustomPopup(
                      context,
                      namespace == Namespace.Organisational
                          ? DomainPopup(
                              parentCallback: callback,
                            )
                          : ObjectPopup(
                              parentCallback: callback, namespace: namespace),
                      isDismissible: true),
                  icon: const Icon(Icons.add),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class SettingsViewPopup extends StatelessWidget {
  final TreeAppController controller;

  const SettingsViewPopup({super.key, required this.controller});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        height: 500,
        child: TreeAppControllerScope(
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
