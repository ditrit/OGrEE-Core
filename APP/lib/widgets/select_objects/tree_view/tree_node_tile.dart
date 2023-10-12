import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/pages/tenant_page.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';

import '../treeapp_controller.dart';
import 'tree_node.dart';

part '_actions_chip.dart';
part '_selector.dart';

const Color _kDarkBlue = Color(0xFF1565C0);

const RoundedRectangleBorder kRoundedRectangleBorder = RoundedRectangleBorder(
  borderRadius: BorderRadius.all(Radius.circular(12)),
);

class TreeNodeTile extends StatefulWidget {
  final bool isTenantMode;
  final TreeEntry<TreeNode> entry;
  final Function() onTap;
  const TreeNodeTile(
      {Key? key,
      required this.isTenantMode,
      required this.entry,
      required this.onTap})
      : super(key: key);

  @override
  _TreeNodeTileState createState() => _TreeNodeTileState();
}

class _TreeNodeTileState extends State<TreeNodeTile> {
  @override
  Widget build(BuildContext context) {
    final appController = TreeAppController.of(context);

    return InkWell(
      onTap: widget.onTap,
      child: TreeIndentation(
        entry: widget.entry,
        guide: const IndentGuide.connectingLines(indent: 48),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(4, 8, 8, 8),
          child: Row(
            children: [
              FolderButton(
                closedIcon: const Icon(Icons.auto_awesome_mosaic),
                openedIcon: const Icon(Icons.auto_awesome_mosaic_outlined),
                icon: widget.isTenantMode
                    ? const Icon(Icons.dns)
                    : const Icon(Icons.auto_awesome_mosaic),
                isOpen:
                    widget.entry.hasChildren ? widget.entry.isExpanded : null,
                onPressed: widget.entry.hasChildren ? widget.onTap : null,
              ),
              _NodeActionsChip(node: widget.entry.node),
              widget.isTenantMode
                  ? Row(
                      children: [
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 8.0),
                          child: CircleAvatar(
                            radius: 13,
                            child: IconButton(
                                splashRadius: 18,
                                iconSize: 14,
                                padding: const EdgeInsets.all(2),
                                onPressed: () => showCustomPopup(
                                    context,
                                    DomainPopup(
                                      parentCallback: () => appController.init(
                                          {},
                                          namespace: Namespace.Organisational,
                                          reload: true,
                                          isTenantMode: true),
                                      domainId: widget.entry.node.id,
                                    )),
                                icon: const Icon(
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
                              padding: const EdgeInsets.all(2),
                              onPressed: () => TenantPage.of(context)!
                                  .changeToUserView(widget.entry.node.id),
                              icon: const Icon(
                                Icons.people,
                                color: Colors.black,
                              )),
                        ),
                      ],
                    )
                  : _NodeSelector(id: widget.entry.node.id),
            ],
          ),
        ),
      ),
    );
  }

  void _describeAncestors(TreeNode node) {
    showSnackBar(
      context,
      '${AppLocalizations.of(context)!.nodePath} ${node.id}',
      duration: const Duration(seconds: 3),
    );
  }
}
