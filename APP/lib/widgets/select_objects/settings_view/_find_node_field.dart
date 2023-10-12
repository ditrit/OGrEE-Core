part of 'settings_view.dart';

class _FindNodeField extends StatefulWidget {
  const _FindNodeField({Key? key}) : super(key: key);

  @override
  __FindNodeFieldState createState() => __FindNodeFieldState();
}

class __FindNodeFieldState extends State<_FindNodeField> {
  late final controller = TextEditingController();

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return TextField(
      controller: controller,
      cursorColor: Colors.blueGrey,
      autofocus: false,
      decoration: GetFormInputDecoration(
        false,
        '${localeMsg.search}...',
        icon: Icons.search_rounded,
      ),
      onSubmitted: (_) => _submitted(),
    );
  }

  void _submitted() {
    final id = controller.text.trim();
    final appController = TreeAppController.of(context);
    final localeMsg = AppLocalizations.of(context)!;

    TreeNode? node;
    for (var root in appController.treeController.roots) {
      if (root.id.toLowerCase().contains(id.toLowerCase())) {
        node = root;
        break;
      }
      node = root.nullableDescendants.firstWhere(
        (descendant) => descendant == null
            ? false
            : descendant.id.toLowerCase().contains(id.toLowerCase()),
        orElse: () => null,
      );
      if (node != null) {
        break;
      }
    }

    if (node == null) {
      showSnackBar(
        context,
        '${localeMsg.noNodeFound} $id',
        duration: const Duration(seconds: 3),
      );
    } else {
      showSnackBar(
        context,
        '${localeMsg.nodeFound} ${node.id}',
        isSuccess: true,
      );
      // Expand only until found node and scroll to it
      if (!appController.treeController.areAllRootsCollapsed) {
        appController.treeController.collapseAll();
      }
      appController.treeController.expandAncestors(node);
      appController.scrollTo(node);
    }
  }
}
