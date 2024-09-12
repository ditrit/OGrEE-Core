part of 'settings_view.dart';

class _AdvancedFindField extends StatefulWidget {
  final Namespace namespace;
  const _AdvancedFindField({required this.namespace});

  @override
  _AdvancedFindFieldState createState() => _AdvancedFindFieldState();
}

class _AdvancedFindFieldState extends State<_AdvancedFindField> {
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
      style: const TextStyle(fontSize: 14),
      decoration: GetFormInputDecoration(
        false,
        localeMsg.expression,
        hint: "name=bladeA&category=device",
        iconWidget: Padding(
          padding: const EdgeInsets.only(right: 12, left: 12.0),
          child: Tooltip(
            message:
                "${localeMsg.advancedSearchHint} (category=device & name=ibm*) | tag=blade-hp",
            verticalOffset: 13,
            decoration: const BoxDecoration(
              color: Colors.blueAccent,
              borderRadius: BorderRadius.all(Radius.circular(12)),
            ),
            textStyle: const TextStyle(
              fontSize: 13,
              color: Colors.white,
            ),
            padding: const EdgeInsets.all(13),
            child: const Icon(
              Icons.info_outline_rounded,
              color: Colors.blueAccent,
            ),
          ),
        ),
      ),
      onSubmitted: (_) => submitted(),
    );
  }

  Future<void> submittedAsFilter() async {
    final searchExpression = controller.text.trim();
    final appController = TreeAppController.of(context);
    final localeMsg = AppLocalizations.of(context)!;
    final messenger = ScaffoldMessenger.of(context);
    List<String> ids;

    if (searchExpression.isEmpty) {
      appController.filterTreeById([]);
      showSnackBar(
        messenger,
        'Filter cleared',
        duration: const Duration(seconds: 3),
        isSuccess: true,
      );
      return;
    }

    final result = await fetchWithComplexFilter(
      searchExpression,
      widget.namespace,
      localeMsg,
    );
    switch (result) {
      case Success(value: final foundObjs):
        ids = getIdsFromObjects(foundObjs);
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        return;
    }

    if (ids.isEmpty) {
      showSnackBar(
        messenger,
        '${localeMsg.noNodeFound} $searchExpression',
        duration: const Duration(seconds: 3),
      );
    } else {
      showSnackBar(
        messenger,
        '${localeMsg.xNodesFound(ids.length)} $searchExpression',
        isSuccess: true,
      );
      appController.filterTreeById(ids);
    }
  }

  List<String> getIdsFromObjects(List<Map<String, dynamic>> foundObjs) {
    final List<String> ids = [];
    for (final obj in foundObjs) {
      ids.add(obj["id"] as String);
    }
    return ids;
  }

  Future<void> submitted() async {
    final searchExpression = controller.text.trim();
    final appController = TreeAppController.of(context);
    final localeMsg = AppLocalizations.of(context)!;
    final messenger = ScaffoldMessenger.of(context);
    List<TreeNode> nodes;

    final result = await fetchWithComplexFilter(
      searchExpression,
      widget.namespace,
      localeMsg,
    );
    switch (result) {
      case Success(value: final foundObjs):
        nodes = getTreeNodesFromObjects(foundObjs, appController);
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        return;
    }

    if (nodes.isEmpty) {
      showSnackBar(
        messenger,
        '${localeMsg.noNodeFound} $searchExpression',
        duration: const Duration(seconds: 3),
      );
    } else {
      showSnackBar(
        messenger,
        '${localeMsg.xNodesFound(nodes.length)} $searchExpression',
        isSuccess: true,
      );
      // Expand only until found nodes and scroll to first one
      if (!appController.treeController.areAllRootsCollapsed) {
        appController.treeController.collapseAll();
      }
      for (final node in nodes) {
        appController.treeController.expandAncestors(node);
        appController.scrollTo(node);
        appController.selectNode(node.id);
      }
      appController.scrollTo(nodes.first);
    }
  }

  List<TreeNode> getTreeNodesFromObjects(
    List<Map<String, dynamic>> foundObjs,
    TreeAppController appController,
  ) {
    final List<TreeNode> nodes = [];
    for (final obj in foundObjs) {
      final id = obj["id"] as String;
      // search for this obj on root node or in its children
      for (final root in appController.treeController.roots) {
        TreeNode? node;
        if (root.id.toLowerCase().contains(id.toLowerCase())) {
          node = root;
        } else {
          node = root.nullableDescendants.firstWhere(
            (descendant) => descendant == null
                ? false
                : descendant.id.toLowerCase().contains(id.toLowerCase()),
            orElse: () => null,
          );
        }
        //found it
        if (node != null) {
          nodes.add(node);
          break;
        }
      }
    }
    return nodes;
  }
}
