import 'package:flutter/material.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'settings_view/settings_view.dart';
import 'tree_view/custom_tree_view.dart';

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
                widget.namespace == 'TEST'
                    ? {}
                    : SelectPage.of(context)!.selectedObjects,
                dateRange: widget.dateRange,
                isTest: widget.namespace == 'TEST',
                reload: widget.load,
                onlyDomain: widget.namespace == 'Organisational')
            : null,
        builder: (_, __) {
          print(widget.load);
          if (appController.isInitialized && widget.load) {
            return _Unfocus(
              child: Card(
                margin: const EdgeInsets.all(0.1),
                child: appController.rootNode.children.length <= 0
                    ? Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.warning_rounded,
                            size: 50,
                            color: Colors.grey.shade600,
                          ),
                          const Padding(
                            padding: EdgeInsets.only(top: 16),
                            child: Text("No objects found :("),
                          ),
                        ],
                      )
                    : _ResponsiveBody(
                        onlyDomain: widget.namespace == 'Organisational'),
                // endDrawer: Drawer(child: SettingsView()),
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
  const _ResponsiveBody({Key? key, this.onlyDomain = false}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    // print("BUILD RespBody");
    if (MediaQuery.of(context).size.width < 600) {
      return const CustomTreeView(isTenantMode: false);
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
