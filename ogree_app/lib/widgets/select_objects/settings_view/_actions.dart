part of 'settings_view.dart';

class _Actions extends StatelessWidget {
  const _Actions({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: [
        _Action(
          label: Text(localeMsg.expandAll),
          onPressed: AppController.of(context).treeController.expandAll,
        ),
        _Action(
          label: Text(localeMsg.collapseAll),
          onPressed: AppController.of(context).treeController.collapseAll,
        ),
        _Action(
          label: Text(localeMsg.selectAll),
          onPressed: AppController.of(context).selectAll,
        ),
        _Action(
          label: Text(localeMsg.deselectAll),
          onPressed: () => AppController.of(context).selectAll(false),
        ),
      ],
    );
  }
}

class _Action extends StatelessWidget {
  const _Action({
    Key? key,
    required this.label,
    this.onPressed,
  }) : super(key: key);

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
