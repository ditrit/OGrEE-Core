part of 'tree_node_tile.dart';

class _NodeSelector extends StatelessWidget {
  final String id;
  const _NodeSelector({required this.id});

  @override
  Widget build(BuildContext context) {
    final appController = TreeAppController.of(context);
    return AnimatedBuilder(
      animation: appController,
      builder: (_, __) {
        return Checkbox(
          shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.all(Radius.circular(3)),
          ),
          activeColor: Colors.green.shade600,
          value: appController.isSelected(id),
          onChanged: (_) => appController.toggleSelection(id),
        );
      },
    );
  }
}
