part of 'tree_node_tile.dart';

class _NodeActionsChip extends StatefulWidget {
  final TreeNode node;
  final bool isVirtual;
  final bool isTemplate;
  const _NodeActionsChip(
      {required this.node,
      this.isTemplate = false,
      this.isVirtual = false,});

  @override
  State<_NodeActionsChip> createState() => _NodeActionsChipState();
}

class _NodeActionsChipState extends State<_NodeActionsChip> {
  final GlobalKey<PopupMenuButtonState> _popupMenuKey = GlobalKey();

  PopupMenuButtonState? get _menu => _popupMenuKey.currentState;

  @override
  Widget build(BuildContext context) {
    final namespace = TreeAppController.of(context).namespace;
    final menuEntries = <PopupMenuEntry<int>>[
      PopupMenuItem(
        value: 1,
        child: ListTile(
          dense: true,
          title: Text(AppLocalizations.of(context)!.toggleSelection),
          contentPadding: const EdgeInsets.symmetric(horizontal: 4),
          leading: const Icon(Icons.account_tree_rounded, color: _kDarkBlue),
        ),
      ),
    ];
    if (namespace != Namespace.Logical || widget.node.id[0] != starSymbol) {
      menuEntries.add(
        PopupMenuItem(
          value: 2,
          child: ListTile(
            dense: true,
            title: Text(AppLocalizations.of(context)!.viewEditNode),
            contentPadding: const EdgeInsets.symmetric(horizontal: 4),
            leading: const Icon(Icons.edit, color: _kDarkBlue),
          ),
        ),
      );
      if (!widget.isTemplate) {
        menuEntries.add(
          PopupMenuItem(
            value: 3,
            child: ListTile(
              dense: true,
              title: Text(AppLocalizations.of(context)!.viewJSON),
              contentPadding: const EdgeInsets.symmetric(horizontal: 4),
              leading: const Icon(Icons.search, color: _kDarkBlue),
            ),
          ),
        );
      }
    }

    if (namespace == Namespace.Physical &&
        '.'.allMatches(widget.node.id).length >= 4) {
      menuEntries.add(
        PopupMenuItem(
          value: 4,
          child: ListTile(
            dense: true,
            title: Text(AppLocalizations.of(context)!.viewGraph),
            contentPadding: const EdgeInsets.symmetric(horizontal: 4),
            leading: const Icon(Icons.auto_graph, color: _kDarkBlue),
          ),
        ),
      );
    }

    return PopupMenuButton<int>(
      key: _popupMenuKey,
      tooltip: AppLocalizations.of(context)!.selectionOptions,
      offset: const Offset(0, 32),
      itemBuilder: (_) => menuEntries,
      onSelected: (int selected) {
        if (selected == 1) {
          TreeAppController.of(context).toggleAllFrom(widget.node);
        } else if (selected == 2) {
          showCustomPopup(
              context,
              namespace == Namespace.Organisational
                  ? DomainPopup(
                      parentCallback: () => TreeAppController.of(context).init(
                          {},
                          argNamespace: Namespace.Organisational,
                          reload: true,
                          isTenantMode: true,),
                      domainId: widget.node.id,
                    )
                  : ObjectPopup(
                      namespace: namespace,
                      parentCallback: () => TreeAppController.of(context).init(
                          {},
                          argNamespace: namespace,
                          reload: true,),
                      objId: widget.node.id,
                    ),
              isDismissible: true,);
        } else if (selected == 3) {
          showCustomPopup(context,
              ViewObjectPopup(namespace: namespace, objId: widget.node.id),
              isDismissible: true,);
        } else {
          showCustomPopup(context, ObjectGraphView(widget.node.id),
              isDismissible: true,);
        }
      },
      child: RawChip(
        onPressed: () => _menu?.showButtonMenu(),
        backgroundColor:
            widget.isVirtual ? Colors.deepPurple.shade100 : const Color(0x331565c0),
        side: const BorderSide(style: BorderStyle.none),
        label: Text(
          adaptLabel(widget.node.label),
          style: TextStyle(
            fontSize: 14,
            fontFamily: GoogleFonts.inter().fontFamily,
            color: widget.isVirtual ? Colors.deepPurple.shade900 : _kDarkBlue,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }

  String adaptLabel(String label) {
    String editedLabel = label;
    if (label.startsWith("*")) {
      editedLabel = label.replaceFirst("*", "").capitalize();
    }
    return editedLabel;
  }
}
