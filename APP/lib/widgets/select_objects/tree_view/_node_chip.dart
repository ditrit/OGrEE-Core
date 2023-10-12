part of 'tree_node_tile.dart';

class _NodeActionsChip extends StatefulWidget {
  final TreeNode node;
  const _NodeActionsChip({Key? key, required this.node}) : super(key: key);

  @override
  State<_NodeActionsChip> createState() => _NodeActionsChipState();
}

class _NodeActionsChipState extends State<_NodeActionsChip> {
  final GlobalKey<PopupMenuButtonState> _popupMenuKey = GlobalKey();

  PopupMenuButtonState? get _menu => _popupMenuKey.currentState;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<int>(
      key: _popupMenuKey,
      tooltip: AppLocalizations.of(context)!.selectionOptions,
      offset: const Offset(0, 32),
      itemBuilder: (_) => <PopupMenuEntry<int>>[
        PopupMenuItem(
          value: 1,
          child: ListTile(
            dense: true,
            title: Text(AppLocalizations.of(context)!.toggleSelection),
            contentPadding: EdgeInsets.symmetric(horizontal: 4),
            leading: Icon(Icons.account_tree_rounded, color: _kDarkBlue),
          ),
        ),
      ],
      onSelected: (int selected) {
        TreeAppController.of(context).toggleAllFrom(widget.node);
      },
      child: RawChip(
        onPressed: () => _menu?.showButtonMenu(),
        backgroundColor: const Color(0x331565c0),
        side: const BorderSide(style: BorderStyle.none),
        label: Text(
          widget.node.label,
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
