part of 'settings_view.dart';

class _FindNodeField extends StatefulWidget {
  const _FindNodeField({Key? key}) : super(key: key);

  @override
  __FindNodeFieldState createState() => __FindNodeFieldState();
}

class __FindNodeFieldState extends State<_FindNodeField> {
  late final controller = TextEditingController();
  late final focusNode = FocusNode();

  @override
  void dispose() {
    controller.dispose();
    focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return TextField(
      controller: controller,
      cursorColor: Colors.blueGrey,
      autofocus: false,
      focusNode: focusNode,
      decoration: InputDecoration(
        isDense: true,
        hintText: '${localeMsg.search}...',
        hintStyle: const TextStyle(
          fontStyle: FontStyle.italic,
        ),
        suffixIcon: IconButton(
          onPressed: _submitted,
          tooltip: localeMsg.search,
          icon: const Icon(
            Icons.search_rounded,
          ),
        ),
      ),
      onSubmitted: (_) => _submitted(),
    );
  }

  void _submitted() {
    final id = controller.text.trim();
    final appController = AppController.of(context);
    final localeMsg = AppLocalizations.of(context)!;

    final node = appController.rootNode.nullableDescendants.firstWhere(
      (descendant) => descendant == null
          ? false
          : descendant.id.toLowerCase().contains(id.toLowerCase()),
      orElse: () => null,
    );

    if (node == null) {
      showSnackBar(
        context,
        '${localeMsg.noNodeFound} $id',
        duration: const Duration(seconds: 3),
      );
    } else {
      if (!appController.treeController.isExpanded(id)) {
        appController.treeController.expandUntil(node);
      }
      appController.scrollTo(node);
    }
    controller.clear();
    focusNode.unfocus();
  }
}
