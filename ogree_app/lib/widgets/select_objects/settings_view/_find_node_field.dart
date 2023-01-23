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
    return TextField(
      controller: controller,
      cursorColor: Colors.blueGrey,
      autofocus: false,
      focusNode: focusNode,
      decoration: InputDecoration(
        isDense: true,
        hintText: 'Search...', // case sensitive
        hintStyle: const TextStyle(
          fontStyle: FontStyle.italic,
        ),
        suffixIcon: IconButton(
          onPressed: _submitted,
          tooltip: 'Search',
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
    final node = appController.treeController.find(id);

    if (node == null) {
      showSnackBar(
        context,
        'No node was found with ID:  $id',
        duration: const Duration(seconds: 3),
      );
    } else {
      appController.toggleSelection(id, shouldSelect: true);
      if (!appController.treeController.isExpanded(id)) {
        appController.treeController.expandUntil(node);
      }
      appController.scrollTo(node);
    }
    controller.clear();
    focusNode.unfocus();
  }
}
