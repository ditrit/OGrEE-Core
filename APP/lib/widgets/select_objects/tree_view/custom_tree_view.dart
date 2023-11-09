import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';

import '../treeapp_controller.dart';
import 'tree_node.dart';
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
    final appController = TreeAppController.of(context);
    return Scrollbar(
      thumbVisibility: false,
      controller: appController.scrollController,
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: SizedBox(
          width: MediaQuery.of(context).size.width > 15
              ? MediaQuery.of(context).size.width - 15
              : MediaQuery.of(context).size.width,
          child: TreeView(
            treeController: appController.treeController,
            controller: appController.scrollController,
            nodeBuilder: (_, TreeEntry<TreeNode> entry) => TreeNodeTile(
              isTenantMode: widget
                  .isTenantMode, // Add a key to your tiles to avoid syncing descendant animations.
              key: ValueKey(entry.node),
              // Tree nodes are wrapped in TreeEntry instances when traversing
              // the tree, these objects hold important details about its node
              // relative to the tree, like: expansion state, level, parent, etc.
              //
              // TreeEntrys are short lived, each time TreeController.rebuild is
              // called, a new TreeEntry is created for each node so its properties
              // are always up to date.
              entry: entry,
              onTap: () =>
                  appController.treeController.toggleExpansion(entry.node),
            ),
          ),
        ),
      ),
    );
  }
}
