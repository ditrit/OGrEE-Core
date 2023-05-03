import 'package:flutter/material.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'settings_view/settings_view.dart';
import 'tree_view/custom_tree_view.dart';

class SelectObjects extends StatefulWidget {
  final String namespace;
  final bool load;

  const SelectObjects({super.key, required this.namespace, required this.load});
  @override
  State<SelectObjects> createState() => _SelectObjectsState();
}

class _SelectObjectsState extends State<SelectObjects> {
  late final AppController appController = AppController();

  @override
  Widget build(BuildContext context) {
    final _isSmallDisplay = MediaQuery.of(context).size.width < 600;
    return AppControllerScope(
      controller: appController,
      child: FutureBuilder<void>(
        future: widget.load
            ? appController.init(
                widget.namespace == 'TEST'
                    ? {}
                    : SelectPage.of(context)!.selectedObjects,
                isTest: widget.namespace == 'TEST')
            : null,
        builder: (_, __) {
          if (appController.isInitialized && widget.load) {
            return const _Unfocus(
              child: Card(
                margin: EdgeInsets.all(0.1),
                child: _ResponsiveBody(),
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
  const _ResponsiveBody({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    // print("BUILD RespBody");
    if (MediaQuery.of(context).size.width < 600) {
      return const CustomTreeView(isTenantMode: false);
    }
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        children: const [
          Flexible(flex: 2, child: CustomTreeView(isTenantMode: false)),
          VerticalDivider(
            width: 1,
            thickness: 1,
            color: Colors.black26,
          ),
          Expanded(child: SettingsView(isTenantMode: false)),
        ],
      ),
    );
  }
}
