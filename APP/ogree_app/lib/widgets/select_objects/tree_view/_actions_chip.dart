part of 'tree_node_tile.dart';

class _NodeActionsChip extends StatefulWidget {
  const _NodeActionsChip({Key? key}) : super(key: key);

  @override
  State<_NodeActionsChip> createState() => _NodeActionsChipState();
}

class _NodeActionsChipState extends State<_NodeActionsChip> {
  final GlobalKey<PopupMenuButtonState> _popupMenuKey = GlobalKey();

  PopupMenuButtonState? get _menu => _popupMenuKey.currentState;

  @override
  Widget build(BuildContext context) {
    final nodeScope = TreeNodeScope.of(context);

    return PopupMenuButton<int>(
      key: _popupMenuKey,
      tooltip: AppLocalizations.of(context)!.selectionOptions,
      offset: const Offset(0, 32),
      itemBuilder: (_) => kPopupMenuItems,
      onSelected: (int selected) {
        // if (selected == 0) {
        //   AppController.of(context).toggleSelection(nodeScope.node.id);
        // } else
        AppController.of(context).toggleAllFrom(nodeScope.node);
      },
      child: RawChip(
        onPressed: () => _menu?.showButtonMenu(),
        backgroundColor: const Color(0x331565c0),
        side: const BorderSide(style: BorderStyle.none),
        label: Text(
          nodeScope.node.label,
          style: TextStyle(
            fontSize: 14,
            fontFamily: GoogleFonts.inter().fontFamily,
            color: _kDarkBlue,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}

const kPopupMenuItems = <PopupMenuEntry<int>>[
  // PopupMenuItem(
  //   value: 0,
  //   // height: 28,
  //   child: ListTile(
  //     dense: true,
  //     title: Text('Inverser ce noeud uniquement'),
  //     // subtitle: Text('Opens dialog to add a child'),
  //     contentPadding: EdgeInsets.symmetric(horizontal: 4),
  //     leading: Icon(Icons.add_box_rounded, color: _kDarkBlue),
  //   ),
  // ),
  // PopupMenuDivider(height: 1),
  PopupMenuItem(
    value: 1,
    // height: 28,
    child: ListTile(
      dense: true,
      title: Text('Inverser sélection du noeud et de tous ses enfants'),
      // subtitle: Text('Moves children one level up'),
      contentPadding: EdgeInsets.symmetric(horizontal: 4),
      leading: Icon(Icons.account_tree_rounded, color: _kDarkBlue),
    ),
  ),
];
