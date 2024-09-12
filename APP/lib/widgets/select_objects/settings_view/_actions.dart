part of 'settings_view.dart';

class _Actions extends StatelessWidget {
  final bool isTenantMode;
  const _Actions({required this.isTenantMode});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    final actionsRow = [
      _Action(
        label: Text(localeMsg.expandAll),
        onPressed: TreeAppController.of(context).treeController.expandAll,
      ),
      _Action(
        label: Text(localeMsg.collapseAll),
        onPressed: TreeAppController.of(context).treeController.collapseAll,
      ),
    ];
    if (!isTenantMode) {
      actionsRow.add(_Action(
        label: Text(localeMsg.selectAll),
        onPressed: TreeAppController.of(context).selectAll,
      ),);
      actionsRow.add(_Action(
        label: Text(localeMsg.deselectAll),
        onPressed: () => TreeAppController.of(context).selectAll(false),
      ),);
    }
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: actionsRow,
    );
  }
}

class _Action extends StatelessWidget {
  const _Action({
    required this.label,
    this.onPressed,
  });

  final Widget label;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton(
      style: OutlinedButton.styleFrom(
        foregroundColor: kDarkBlue,
        backgroundColor: const Color(0x331565c0),
        padding: const EdgeInsets.all(10),
        side: const BorderSide(style: BorderStyle.none),
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
        ),
        textStyle: TextStyle(
          fontSize: 13.5,
          fontFamily: GoogleFonts.inter().fontFamily,
          color: kDarkBlue,
          fontWeight: FontWeight.w700,
        ),
      ),
      onPressed: onPressed,
      child: label,
    );
  }
}
