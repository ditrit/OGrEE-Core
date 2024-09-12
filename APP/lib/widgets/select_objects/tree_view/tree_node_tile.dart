import 'package:flutter/material.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/tenant_page.dart';
import 'package:ogree_app/widgets/object_graph_view.dart';
import 'package:ogree_app/widgets/select_objects/object_popup.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/select_objects/tree_view/tree_node.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';
import 'package:ogree_app/widgets/select_objects/view_object_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';

part '_node_chip.dart';
part '_node_selector.dart';

const Color _kDarkBlue = Color(0xFF1565C0);

const RoundedRectangleBorder kRoundedRectangleBorder = RoundedRectangleBorder(
  borderRadius: BorderRadius.all(Radius.circular(12)),
);

class TreeNodeTile extends StatefulWidget {
  final bool isTenantMode;
  final TreeEntry<TreeNode> entry;
  final Function() onTap;
  const TreeNodeTile({
    super.key,
    required this.isTenantMode,
    required this.entry,
    required this.onTap,
  });

  @override
  State<TreeNodeTile> createState() => _TreeNodeTileState();
}

class _TreeNodeTileState extends State<TreeNodeTile> {
  @override
  Widget build(BuildContext context) {
    final appController = TreeAppController.of(context);

    bool isVirtual = false;
    bool isTemplate = false;
    if (appController.fetchedCategories["virtual_obj"] != null &&
        appController.fetchedCategories["virtual_obj"]!
            .contains(widget.entry.node.id)) {
      isVirtual = true;
    } else if (widget.entry.parent != null &&
        widget.entry.parent!.node.id.contains("template")) {
      isTemplate = true;
    }

    return InkWell(
      hoverColor: Colors.white,
      onTap: widget.onTap,
      onDoubleTap: () => _describeAncestors(widget.entry.node),
      child: TreeIndentation(
        entry: widget.entry,
        guide: const IndentGuide.connectingLines(indent: 48),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(4, 8, 8, 8),
          child: Row(
            children: [
              FolderButton(
                closedIcon: isVirtual
                    ? const Icon(Icons.cloud)
                    : const Icon(Icons.auto_awesome_mosaic),
                openedIcon: isVirtual
                    ? const Icon(Icons.cloud_outlined)
                    : const Icon(Icons.auto_awesome_mosaic_outlined),
                icon: widget.isTenantMode
                    ? const Icon(Icons.dns)
                    : isVirtual
                        ? const Icon(Icons.cloud)
                        : const Icon(Icons.auto_awesome_mosaic),
                isOpen:
                    widget.entry.hasChildren ? widget.entry.isExpanded : null,
                onPressed: widget.entry.hasChildren ? widget.onTap : null,
              ),
              _NodeActionsChip(
                node: widget.entry.node,
                isTemplate: isTemplate,
                isVirtual: isVirtual,
              ),
              if (widget.isTenantMode)
                Row(
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
                                argNamespace: Namespace.Organisational,
                                reload: true,
                                isTenantMode: true,
                              ),
                              domainId: widget.entry.node.id,
                            ),
                          ),
                          icon: const Icon(
                            Icons.edit,
                            color: Colors.black,
                          ),
                        ),
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
                        ),
                      ),
                    ),
                  ],
                )
              else
                Row(
                  children: [
                    if (widget.entry.node.id[0] == starSymbol)
                      Container()
                    else
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 4.0),
                        child: _NodeSelector(id: widget.entry.node.id),
                      ),
                    if (TreeAppController.of(context).namespace !=
                        Namespace.Logical)
                      CircleAvatar(
                        radius: 10,
                        child: IconButton(
                          splashRadius: 18,
                          iconSize: 14,
                          padding: const EdgeInsets.all(2),
                          onPressed: () => showCustomPopup(
                            context,
                            TreeAppController.of(context).namespace ==
                                    Namespace.Organisational
                                ? DomainPopup(
                                    parentCallback: () => TreeAppController.of(
                                      context,
                                    ).init(
                                      {},
                                      argNamespace: Namespace.Organisational,
                                      reload: true,
                                      isTenantMode: true,
                                    ),
                                    parentId: widget.entry.node.id,
                                  )
                                : ObjectPopup(
                                    namespace: TreeAppController.of(
                                      context,
                                    ).namespace,
                                    parentCallback: () => appController.init(
                                      {},
                                      argNamespace:
                                          TreeAppController.of(context)
                                              .namespace,
                                      reload: true,
                                      isTenantMode: true,
                                    ),
                                    parentId: widget.entry.node.id,
                                  ),
                            isDismissible: true,
                          ),
                          icon: const Icon(
                            Icons.add,
                            color: Colors.black,
                          ),
                        ),
                      )
                    else
                      Container(),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }

  void _describeAncestors(TreeNode node) {
    showSnackBar(
      ScaffoldMessenger.of(context),
      '${AppLocalizations.of(context)!.nodePath} ${node.id}',
      duration: const Duration(seconds: 3),
    );
  }
}
