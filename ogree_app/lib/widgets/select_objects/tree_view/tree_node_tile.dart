import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:ogree_app/common/snackbar.dart';

import '../app_controller.dart';

part '_actions_chip.dart';
part '_selector.dart';

const Color _kDarkBlue = Color(0xFF1565C0);

const RoundedRectangleBorder kRoundedRectangleBorder = RoundedRectangleBorder(
  borderRadius: BorderRadius.all(Radius.circular(12)),
);

class TreeNodeTile extends StatefulWidget {
  const TreeNodeTile({Key? key}) : super(key: key);

  @override
  _TreeNodeTileState createState() => _TreeNodeTileState();
}

class _TreeNodeTileState extends State<TreeNodeTile> {
  @override
  Widget build(BuildContext context) {
    final appController = AppController.of(context);
    final nodeScope = TreeNodeScope.of(context);

    return InkWell(
        onTap: () => _describeAncestors(nodeScope.node),
        onLongPress: () => appController.toggleSelection(nodeScope.node.id),
        child: Row(children: const [
          LinesWidget(),
          NodeWidgetLeadingIcon(
            expandIcon: Icon(Icons.auto_awesome_mosaic),
            collapseIcon: Icon(Icons.auto_awesome_mosaic_outlined),
            leafIcon: Icon(Icons.dns),
          ),
          _NodeActionsChip(),
          _NodeSelector(),
        ]));
  }

  void _describeAncestors(TreeNode node) {
    showSnackBar(
      context,
      'Chemin du noeud : ${node.id}',
      duration: const Duration(seconds: 3),
    );
  }
}
