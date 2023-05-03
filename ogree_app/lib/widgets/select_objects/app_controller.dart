import 'package:flutter/widgets.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:ogree_app/common/api_backend.dart';

class AppController with ChangeNotifier {
  bool _isInitialized = false;
  bool get isInitialized => _isInitialized;
  Map<String, List<String>> fetchedData = {};
  Map<String, List<String>> fetchedCategories = {};
  final Map<int, List<String>> _filterLevels = {};
  static const lastFilterLevel = 3;

  static AppController of(BuildContext context) {
    return context
        .dependOnInheritedWidgetOfExactType<AppControllerScope>()!
        .controller;
  }

  Future<void> init(Map<String, bool> nodes,
      {bool isTest = false,
      bool onlyDomain = false,
      bool reload = false}) async {
    print("INIIIIIIIIIT");
    if (_isInitialized && !reload) return;
    print("FOR REAL INIIIIIIIIIT");
    final rootNode = TreeNode(id: kRootId);
    print("Get API data");
    if (onlyDomain) {
      fetchedData = (await fetchObjectsTree(onlyDomain: true)).first;
      print(fetchedData);
    } else if (isTest) {
      fetchedData = kDataSample;
      fetchedCategories = kDataSampleCategories;
    } else {
      var resp = await fetchObjectsTree();
      fetchedData = resp[0];
      fetchedCategories = resp[1];
    }

    if (reload) {
      // Regenerate tree
      treeController.rootNode
          .clearChildren()
          .forEach((child) => child.delete(recursive: true));
      generateTree(treeController.rootNode, fetchedData);
      // Force redraw tree view
      treeController.refreshNode(treeController.rootNode);
      treeController.reset();
    } else {
      generateTree(rootNode, fetchedData);

      treeController = TreeViewController(
        rootNode: rootNode,
      );
      _isInitialized = true;
      selectedNodes = nodes;
      for (var i = 0; i <= lastFilterLevel; i++) {
        _filterLevels[i] = [];
      }
    }
  }

  //* == == == == == TreeView == == == == ==

  late Map<String, bool> selectedNodes;

  bool isSelected(String id) => selectedNodes[id] ?? false;

  void toggleSelection(String id,
      {bool? shouldSelect, bool shouldNotify = true}) {
    shouldSelect ??= !isSelected(id);
    shouldSelect ? _select(id) : _deselect(id);

    if (shouldNotify) notifyListeners();
  }

  void _select(String id) => selectedNodes[id] = true;

  void _deselect(String id) => selectedNodes.remove(id);

  void selectAll([bool select = true]) {
    if (select) {
      rootNode.descendants.forEach(
        (descendant) => selectedNodes[descendant.id] = true,
      );
    } else {
      rootNode.descendants.forEach(
        (descendant) => selectedNodes.remove(descendant.id),
      );
    }
    notifyListeners();
  }

  void toggleAllFrom(TreeNode node) {
    toggleSelection(node.id);
    node.descendants.forEach(
      (descendant) => toggleSelection(descendant.id),
    );
    notifyListeners();
  }

  void filterTree(String id, int level) {
    // Deep copy original data
    Map<String, List<String>> filteredData = {};
    for (var item in fetchedData.keys) {
      filteredData[item] = List<String>.from(fetchedData[item]!);
    }

    // Add or remove filter
    if (level < 0) {
      // Clear All
      for (var level in _filterLevels.keys) {
        _filterLevels[level] = [];
      }
    } else {
      var currentLevel = _filterLevels[level]!;
      if (!currentLevel.contains(id)) {
        currentLevel.add(id);
      } else {
        currentLevel.remove(id);
      }
    }

    for (var i = 0; i <= lastFilterLevel; i++) {
      if (_filterLevels[i]!.isNotEmpty) {
        filteredData[kRootId] = _filterLevels[i]!;
        break;
      }
    }

    print("FILTERS");
    print(_filterLevels);

    // Find root filter level
    var testLevel = lastFilterLevel;
    List<String> filters = List<String>.from(_filterLevels[testLevel]!);
    while (filters.isEmpty && testLevel > 0) {
      testLevel--;
      filters = List<String>.from(_filterLevels[testLevel]!);
    }
    // Apply all filters from root and bellow
    while (testLevel > 0) {
      List<String> newList = [];
      for (var i = 0; i < filters.length; i++) {
        var parent =
            filters[i].substring(0, filters[i].lastIndexOf('.')); //parent
        filteredData[parent]!.removeWhere((element) {
          return !filters.contains(element);
        });
        newList.add(parent);
      }
      filters = newList;
      testLevel--;
    }

    // Regenerate tree
    treeController.rootNode
        .clearChildren()
        .forEach((child) => child.delete(recursive: true));
    generateTree(treeController.rootNode, filteredData);
    // Force redraw tree view
    treeController.refreshNode(treeController.rootNode);
    treeController.reset();
  }

  TreeNode get rootNode => treeController.rootNode;

  late final TreeViewController treeController;

  //* == == == == == Scroll == == == == ==

  final nodeHeight = 50.0;

  late final scrollController = ScrollController();

  void scrollTo(TreeNode node) {
    final nodeToScroll = node.parent == rootNode ? node : node.parent ?? node;
    final offset = treeController.indexOf(nodeToScroll) * nodeHeight;

    scrollController.animateTo(
      offset,
      duration: const Duration(milliseconds: 500),
      curve: Curves.ease,
    );
  }

  //* == == == == == General == == == == ==

  final treeViewTheme =
      ValueNotifier(const TreeViewTheme(roundLineCorners: true, indent: 64));

  void updateTheme(TreeViewTheme theme) {
    treeViewTheme.value = theme;
  }

  @override
  void dispose() {
    treeController.dispose();
    scrollController.dispose();

    treeViewTheme.dispose();
    super.dispose();
  }
}

class AppControllerScope extends InheritedWidget {
  const AppControllerScope({
    Key? key,
    required this.controller,
    required Widget child,
  }) : super(key: key, child: child);

  final AppController controller;

  @override
  bool updateShouldNotify(AppControllerScope oldWidget) => false;
}

void generateTree(TreeNode parent, Map<String, List<String>> data) {
  final childrenIds = data[parent.id];
  if (childrenIds == null) return;

  parent.addChildren(
    childrenIds.map(
      (String childId) => TreeNode(
          id: childId,
          label: parent.id == kRootId
              ? childId
              : childId.substring(childId.lastIndexOf(".") + 1)),
    ),
  );
  for (var node in parent.children) {
    generateTree(node, data);
  }
}

const String kRootId = 'Root';

const Map<String, List<String>> kDataSample = {
  kRootId: ['sitePA', 'sitePI', 'siteNO', 'sitePB'],
  'sitePA': ['sitePA.A1', 'sitePA.A2'],
  'sitePA.A2': ['sitePA.A2.1'],
  'sitePI': ['sitePI.B1', 'sitePI.B2', 'sitePI.B3'],
  'sitePI.B1': ['sitePI.B1.1', 'sitePI.B1.2', 'sitePI.B1.3'],
  'sitePI.B1.1': ['sitePI.B1.1.rack1', 'sitePI.B1.1.rack2'],
  'sitePI.B1.1.rack1': [
    'sitePI.B1.1.rack1.devA',
    'sitePI.B1.1.rack1.devB',
    'sitePI.B1.1.rack1.devC',
    'sitePI.B1.1.rack1.devD'
  ],
  'sitePI.B1.1.rack2': [
    'sitePI.B1.1.rack2.devA',
    'sitePI.B1.1.rack2.devB',
    'sitePI.B1.1.rack2.devC',
    'sitePI.B1.1.rack2.devD'
  ],
  'sitePI.B1.1.rack2.devB': [
    'sitePI.B1.1.rack2.devB.1',
    'sitePI.B1.1.rack2.devB.devB-2'
  ],
  'sitePI.B1.1.rack2.devC': [
    'sitePI.B1.1.rack2.devC.1',
    'sitePI.B1.1.rack2.devC.devC-2'
  ],
  'sitePI.B2': ['sitePI.B2.1'],
  'sitePI.B2.1': ['sitePI.B2.1.rack1'],
  'siteNO': ['siteNO.BA1', 'siteNO.BB1', 'siteNO.BI1', 'siteNO.BL']
};

const Map<String, List<String>> kDataSampleCategories = {
  "KeysOrder": ["site", "building", "room", "rack"],
  "site": ['sitePA', 'sitePI', 'siteNO', 'sitePB'],
  "building": [
    'sitePA.A1',
    'sitePA.A2',
    'sitePI.B1',
    'sitePI.B2',
    'sitePI.B3',
    'siteNO.BA1',
    'siteNO.BB1',
    'siteNO.BI1',
    'siteNO.BL'
  ],
  "room": [
    'sitePA.A2.1',
    'sitePI.B1.1',
    'sitePI.B1.2',
    'sitePI.B1.3',
    'sitePI.B2.1'
  ],
  "rack": ['sitePI.B1.1.rack1', 'sitePI.B1.1.rack2', 'sitePI.B2.1.rack1']
};
