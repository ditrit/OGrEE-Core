part of 'settings_view.dart';

class SelectedChips extends StatefulWidget {
  const SelectedChips({super.key});

  @override
  State<SelectedChips> createState() => _SelectedChipsState();
}

class _SelectedChipsState extends State<SelectedChips> {
  // Group by
  Map<String, bool> shouldGroupBy = {};

  @override
  Widget build(BuildContext context) {
    final appController = TreeAppController.of(context);
    if (shouldGroupBy.isEmpty &&
        appController.fetchedCategories["room"] != null) {
      for (final room in appController.fetchedCategories["room"]!) {
        shouldGroupBy[room] = true;
      }
    }
    return AnimatedBuilder(
        animation: appController,
        builder: (_, __) {
          return Wrap(
            spacing: 10,
            runSpacing: 10,
            children:
                getChips(Map.from(appController.selectedNodes), appController),
          );
        },);
  }

  List<Widget> getChips(Map<String, bool> nodes, appController) {
    final List<Widget> chips = [];
    final Map<String, List<String>> groups = {}; // group name: [group nodes]

    // Group by, create groups
    for (final key in nodes.keys) {
      for (final group in shouldGroupBy.keys) {
        if (key.contains(group)) {
          if (!groups.containsKey(group)) {
            groups[group] = [key];
          } else {
            groups[group]!.add(key);
          }
          nodes[key] = !shouldGroupBy[group]!;
        }
      }
    }

    // Group chips
    groups.forEach((key, value) {
      if (value.length > 5) {
        chips.add(RawChip(
          onPressed: () => setState(() {
            shouldGroupBy[key] = !shouldGroupBy[key]!;
          }),
          side: const BorderSide(style: BorderStyle.none),
          backgroundColor: Colors.blue.shade200,
          tooltip: value.reduce((value, element) => value = '$value\n$element'),
          label: Text(
            shouldGroupBy[key]! ? "(${value.length}) $key..." : "$key...",
            style: TextStyle(
              fontSize: 14,
              fontFamily: GoogleFonts.inter().fontFamily,
              color: Colors.blue.shade900,
              fontWeight: FontWeight.w600,
            ),
          ),
          avatar: Icon(
            shouldGroupBy[key]!
                ? Icons.group_work_rounded
                : Icons.group_work_outlined,
            size: 20,
            color: Colors.blue.shade900,
          ),
        ),);
      } else {
        for (final element in value) {
          nodes[element] = true;
        }
      }
    });

    // Single chips
    nodes.forEach((key, value) {
      if (value) {
        chips.add(RawChip(
          onPressed: () => appController.toggleSelection(key),
          backgroundColor: Colors.lightGreen.shade200,
          side: const BorderSide(style: BorderStyle.none),
          label: Text(
            key,
            style: TextStyle(
              fontSize: 14,
              fontFamily: GoogleFonts.inter().fontFamily,
              color: Colors.green.shade900,
              fontWeight: FontWeight.w600,
            ),
          ),
          avatar: Icon(
            Icons.cancel,
            size: 20,
            color: Colors.green.shade900,
          ),
        ),);
      }
    });
    return chips;
  }
}
