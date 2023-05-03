import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';

import '../app_controller.dart';
import 'tree_node_tile.dart';

class CustomTreeView extends StatefulWidget {
  final bool isTenantMode;
  const CustomTreeView({Key? key, required this.isTenantMode})
      : super(key: key);

  @override
  _CustomTreeViewState createState() => _CustomTreeViewState();
}

class _CustomTreeViewState extends State<CustomTreeView> {
  @override
  Widget build(BuildContext context) {
    final appController = AppController.of(context);

    return ValueListenableBuilder<TreeViewTheme>(
      valueListenable: appController.treeViewTheme,
      builder: (_, treeViewTheme, __) {
        return Scrollbar(
          thumbVisibility: false,
          controller: appController.scrollController,
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: SizedBox(
              width: MediaQuery.of(context).size.width - 15,
              child: TreeView(
                controller: appController.treeController,
                theme: treeViewTheme,
                scrollController: appController.scrollController,
                nodeHeight: appController.nodeHeight,
                nodeBuilder: (_, __) =>
                    TreeNodeTile(isTenantMode: widget.isTenantMode),
              ),
            ),
          ),
        );
      },
    );
  }
}
