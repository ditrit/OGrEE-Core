// This class represents one node in the Tree.
class TreeNode {
  /// Creates a [TreeNode].
  ///
  /// Use [id] to dynamically manage this node later.
  TreeNode({required this.id, this.label = ''});

  final String id;
  final String label;

  final Set<TreeNode> _children = {};
  Set<TreeNode> get children => _children;
  bool get hasChildren => _children.isNotEmpty;

  /// Convenience operator to get the [index]th child.
  TreeNode operator [](int index) => _children.elementAt(index);

  /// Returns an [Iterable] of every [TreeNode] under this.
  Iterable<TreeNode> get descendants sync* {
    for (final child in _children) {
      yield child;

      if (child.hasChildren) {
        yield* child.descendants;
      }
    }
  }

  /// Same as [descendants] but with nullable return, useful when
  /// filtering nodes to use `orElse: () => null` when no node was found.
  Iterable<TreeNode?> get nullableDescendants sync* {
    for (final child in _children) {
      yield child;
      if (child.hasChildren) {
        yield* child.nullableDescendants;
      }
    }
  }

  TreeNode? get lastChild => _children.isEmpty ? null : _children.last;

  /// Adds a single child to this node and sets its [parent] property to `this`.
  /// If [child]'s `parent != null`, it will be removed from the children of
  /// it's old parent before being added to this.
  void addChild(TreeNode child) {
    // A node can't be neither child of its children nor parent of itself.
    if (child == parent || child == this) return;

    child.parent?.removeChild(child);
    child._parent = this;
    _children.add(child);
  }

  /// Adds a list of children to this node.
  void addChildren(Iterable<TreeNode> nodes) => nodes.forEach(addChild);

  /// Removes a single child from this node and set its parent to `null`.
  void removeChild(TreeNode child) {
    final wasRemoved = _children.remove(child);

    if (wasRemoved) {
      child._parent = null;
    }
  }

  /// Removes this node from the tree.
  ///
  /// Moves every child in [this.children] to [this.parent.children] and
  /// removes [this] from [this.parent.children].
  ///
  /// Example:
  /// ```
  /// /*
  /// rootNode
  ///   |-- childNode1
  ///   │     |-- grandChildNode1
  ///   │     '-- grandChildNode2
  ///   '-- childNode2
  ///
  /// childNode1.delete() is called, the tree becomes:
  ///
  /// rootNode
  ///   |-- childNode2
  ///   |-- grandChildNode1
  ///   '-- grandChildNode2
  /// */
  /// ```
  /// Set `recursive` to `true` if you want to delete the entire subtree.
  /// (Ps: if the subtree is too large, this might take a while.)
  ///
  /// If [parent] is null, this method has no effects.
  void delete({bool recursive = false}) {
    if (isRoot) return;

    if (recursive) {
      clearChildren().forEach((child) => child.delete(recursive: true));
    } else {
      _parent!.addChildren(clearChildren());
    }
    _parent!.removeChild(this);
  }

  // Removes all children from this node and sets their parent to null.
  // Returns the old children to easily move nodes to another parent.
  List<TreeNode> clearChildren() {
    final removedChildren = _children.map((child) {
      child._parent = null;
      return child;
    }).toList(growable: false);

    _children.clear();
    return removedChildren;
  }

  // If `null`, this node is the root of the tree
  // or it doesn't belong to any node yet.
  // This property is set by [TreeNode.addChild].
  TreeNode? get parent => _parent;
  TreeNode? _parent;

  // Returns the path from the root node to this node, not including this.
  Iterable<TreeNode> get ancestors sync* {
    if (parent != null) {
      yield* parent!.ancestors;
      yield parent!;
    }
  }

  int get depth => isRoot ? -1 : parent!.depth + 1;
  bool get isLeaf => _children.isEmpty;
  bool get isRoot => parent == null;

  /// Whether this node is a direct child of the root node.
  bool get isMostTopLevel => depth == 0;

  // Whether or not this node is the last child of its parent.
  // If this method throws, the tree was malformed.
  bool get hasNextSibling => isRoot ? false : this != parent!.lastChild;

  // Starting from this node, searches the subtree
  // looking for a node id that match [id],
  // returns `null` if no node was found with the given [id].
  TreeNode? find(String id) => nullableDescendants.firstWhere(
        (descendant) => descendant == null ? false : descendant.id == id,
        orElse: () => null,
      );

  int compareTo(TreeNode other) => id.compareTo(other.id);

  @override
  String toString() => 'TreeNode(id: $id, label: $label)';
}
