import 'package:flutter/widgets.dart';
import 'package:flutter_fancy_tree_view/flutter_fancy_tree_view.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/theme.dart';

import 'package:ogree_app/widgets/select_objects/tree_view/tree_node.dart';

bool isSmallDisplay = false;

class TreeAppController with ChangeNotifier {
  bool _isInitialized = false;
  bool get isInitialized => _isInitialized;
  Namespace namespace = Namespace.Physical;

  Map<String, List<String>> fetchedData = {};
  Map<String, List<String>> fetchedCategories = {};

  final Map<int, List<String>> _filterLevels = {};
  Map<int, List<String>> get filterLevels => _filterLevels;
  static const lastFilterLevel = 3;

  late final TreeController<TreeNode> treeController;

  static TreeAppController of(BuildContext context) {
    isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return context
        .dependOnInheritedWidgetOfExactType<TreeAppControllerScope>()!
        .controller;
  }

  Future<void> init(
    Map<String, bool> nodes, {
    Namespace argNamespace = Namespace.Physical,
    bool reload = false,
    String dateRange = "",
    bool isTenantMode = false,
  }) async {
    if (_isInitialized && !reload) return;
    final rootNode = TreeNode(id: kRootId);

    // Fetch data for the tree
    if (argNamespace == Namespace.Test) {
      fetchedData = kDataSample;
      fetchedCategories = kDataSampleCategories;
    } else {
      namespace = argNamespace;
      final result = await fetchObjectsTree(
        dateRange: dateRange,
        namespace: argNamespace,
        isTenantMode: isTenantMode,
      );

      switch (result) {
        case Success(value: final listValue):
          fetchedData = listValue[0];
          fetchedCategories = listValue[1];
        case Failure():
      }
    }

    generateTree(rootNode, fetchedData);

    if (_isInitialized && reload) {
      // Rebuild tree
      treeController.roots = rootNode.children;
      treeController.rebuild();
    } else {
      // Build tree
      treeController = TreeController<TreeNode>(
        roots: rootNode.children,
        childrenProvider: (TreeNode node) => node.children,
        parentProvider: (TreeNode node) => node.parent,
      );
      _isInitialized = true;
      selectedNodes = nodes;
      for (var i = 0; i <= lastFilterLevel; i++) {
        _filterLevels[i] = [];
      }
    }
  }

  deepCopy(Map<String, List<String>> source, destination) {
    for (final item in source.keys) {
      destination[item] = List<String>.from(source[item]!);
    }
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
              : childId.substring(childId.lastIndexOf(".") + 1),
        ),
      ),
    );
    for (final node in parent.children) {
      generateTree(node, data);
    }
  }

  // Tree Node Selection Functionality
  late Map<String, bool> selectedNodes;
  bool isSelected(String id) => selectedNodes[id] ?? false;

  void toggleSelection(
    String id, {
    bool? shouldSelect,
    bool shouldNotify = true,
  }) {
    shouldSelect ??= !isSelected(id);
    shouldSelect ? select(id) : deselect(id);

    if (shouldNotify) notifyListeners();
  }

  void selectAll([bool select = true]) {
    //treeController.expandAll();
    if (select) {
      for (final root in treeController.roots) {
        if (root.id[0] != starSymbol) {
          selectedNodes[root.id] = true;
        }
        for (final descendant in root.descendants) {
          selectedNodes[descendant.id] = true;
        }
      }
    } else {
      for (final root in treeController.roots) {
        selectedNodes.remove(root.id);
        for (final descendant in root.descendants) {
          selectedNodes.remove(descendant.id);
        }
      }
    }
    notifyListeners();
  }

  void selectNode(String id) {
    select(id);
    notifyListeners();
  }

  void deselectNode(String id) {
    selectedNodes.remove(id);
    notifyListeners();
  }

  void select(String id) => selectedNodes[id] = true;
  void deselect(String id) => selectedNodes.remove(id);

  void toggleAllFrom(TreeNode node) {
    if (node.id[0] != starSymbol) {
      toggleSelection(node.id);
    }
    for (final descendant in node.descendants) {
      toggleSelection(descendant.id);
    }
    notifyListeners();
  }

  // Filter Tree Functionality
  void filterTree(String id, int level) {
    // Deep copy original data
    final Map<String, List<String>> filteredData = {};
    deepCopy(fetchedData, filteredData);

    // Add or remove filter
    if (level < 0) {
      // Clear All
      for (final level in _filterLevels.keys) {
        _filterLevels[level] = [];
      }
    } else {
      final currentLevel = _filterLevels[level]!;
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

    // Find root filter level
    var testLevel = lastFilterLevel;
    List<String> filters = List<String>.from(_filterLevels[testLevel]!);
    while (filters.isEmpty && testLevel > 0) {
      testLevel--;
      filters = List<String>.from(_filterLevels[testLevel]!);
    }
    // Apply all filters from root and bellow
    while (testLevel > 0) {
      final List<String> newList = [];
      for (var i = 0; i < filters.length; i++) {
        final parent =
            filters[i].substring(0, filters[i].lastIndexOf('.')); //parent
        if (filteredData[parent] != null) {
          filteredData[parent]!.removeWhere((element) {
            return !filters.contains(element);
          });
          newList.add(parent);
        }
      }
      filters = newList;
      testLevel--;
    }

    // Regenerate tree
    final rootNode = TreeNode(id: kRootId);
    generateTree(rootNode, filteredData);
    treeController.roots = rootNode.children;
    treeController.rebuild();
  }

  filterTreeById(List<String> ids) {
    final Map<String, List<String>> filteredData = {};
    if (ids.isEmpty) {
      for (final item in fetchedData.keys) {
        filteredData[item] = List<String>.from(fetchedData[item]!);
      }
    } else {
      filteredData[kRootId] = [];
      for (final id in ids) {
        filteredData[kRootId]!.add(id);
      }
    }

    // Regenerate tree
    final rootNode = TreeNode(id: kRootId);
    generateTree(rootNode, filteredData);
    treeController.roots = rootNode.children;
    treeController.rebuild();
  }

  // Tree Scroll Functionality
  final nodeHeight = 50.0;
  late final scrollController = ScrollController();
  void scrollTo(TreeNode node) {
    var offset = node.depth * nodeHeight;
    if (node.ancestors.isNotEmpty) {
      final parent = node.ancestors.last;
      offset = offset + parent.children.toList().indexOf(node) * nodeHeight;
    }
    scrollController.animateTo(
      offset,
      duration: const Duration(milliseconds: 500),
      curve: Curves.ease,
    );
  }

  @override
  void dispose() {
    treeController.dispose();
    scrollController.dispose();
    super.dispose();
  }
}

class TreeAppControllerScope extends InheritedWidget {
  const TreeAppControllerScope({
    super.key,
    required this.controller,
    required super.child,
  });

  final TreeAppController controller;

  @override
  bool updateShouldNotify(TreeAppControllerScope oldWidget) => false;
}

const String kRootId = '*';

// Sample and test data
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
    'sitePI.B1.1.rack1.devD',
  ],
  'sitePI.B1.1.rack2': [
    'sitePI.B1.1.rack2.devA',
    'sitePI.B1.1.rack2.devB',
    'sitePI.B1.1.rack2.devC',
    'sitePI.B1.1.rack2.devD',
  ],
  'sitePI.B1.1.rack2.devB': [
    'sitePI.B1.1.rack2.devB.1',
    'sitePI.B1.1.rack2.devB.devB-2',
  ],
  'sitePI.B1.1.rack2.devC': [
    'sitePI.B1.1.rack2.devC.1',
    'sitePI.B1.1.rack2.devC.devC-2',
  ],
  'sitePI.B2': ['sitePI.B2.1'],
  'sitePI.B2.1': ['sitePI.B2.1.rack1'],
  'siteNO': ['siteNO.BA1', 'siteNO.BB1', 'siteNO.BI1', 'siteNO.BL'],
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
    'siteNO.BL',
  ],
  "room": [
    'sitePA.A2.1',
    'sitePI.B1.1',
    'sitePI.B1.2',
    'sitePI.B1.3',
    'sitePI.B2.1',
  ],
  "rack": ['sitePI.B1.1.rack1', 'sitePI.B1.1.rack2', 'sitePI.B2.1.rack1'],
};
