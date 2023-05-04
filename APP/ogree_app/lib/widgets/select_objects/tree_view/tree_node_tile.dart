import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';

import '../app_controller.dart';

part '_actions_chip.dart';
part '_selector.dart';

const Color _kDarkBlue = Color(0xFF1565C0);

const RoundedRectangleBorder kRoundedRectangleBorder = RoundedRectangleBorder(
  borderRadius: BorderRadius.all(Radius.circular(12)),
);

class TreeNodeTile extends StatefulWidget {
  final bool isTenantMode;
  const TreeNodeTile({Key? key, required this.isTenantMode}) : super(key: key);

  @override
  _TreeNodeTileState createState() => _TreeNodeTileState();
}

class _TreeNodeTileState extends State<TreeNodeTile> {
  @override
  Widget build(BuildContext context) {
    final appController = AppController.of(context);
    final nodeScope = TreeNodeScope.of(context);

    return InkWell(
        hoverColor: Colors.white,
        onTap: () => _describeAncestors(nodeScope.node),
        onLongPress: () => appController.toggleSelection(nodeScope.node.id),
        child: Row(children: [
          const LinesWidget(),
          NodeWidgetLeadingIcon(
            expandIcon: const Icon(Icons.auto_awesome_mosaic),
            collapseIcon: const Icon(Icons.auto_awesome_mosaic_outlined),
            leafIcon: widget.isTenantMode
                ? const Icon(Icons.dns)
                : const Icon(Icons.auto_awesome_mosaic),
          ),
          const _NodeActionsChip(),
          widget.isTenantMode
              ? Row(
                  children: [
                    Padding(
                      padding: EdgeInsets.symmetric(horizontal: 8.0),
                      child: CircleAvatar(
                        radius: 13,
                        child: IconButton(
                            splashRadius: 18,
                            iconSize: 14,
                            padding: EdgeInsets.all(2),
                            onPressed: () => showCustomPopup(
                                context,
                                DomainPopup(
                                  parentCallback: () => appController
                                      .init({}, onlyDomain: true, reload: true),
                                  domainId: nodeScope.node.id,
                                )),
                            icon: Icon(
                              Icons.edit,
                              color: Colors.black,
                            )),
                      ),
                    ),
                    CircleAvatar(
                      radius: 13,
                      child: IconButton(
                          splashRadius: 18,
                          iconSize: 14,
                          padding: EdgeInsets.all(2),
                          onPressed: null,
                          icon: Icon(
                            Icons.people,
                            color: Colors.black,
                          )),
                    ),
                  ],
                )
              : const _NodeSelector(),
        ]));
  }

  void _describeAncestors(TreeNode node) {
    showSnackBar(
      context,
      '${AppLocalizations.of(context)!.nodePath} ${node.id}',
      duration: const Duration(seconds: 3),
    );
  }
}
