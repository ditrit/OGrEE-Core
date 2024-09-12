import 'package:flutter/material.dart';
import 'package:graphview/GraphView.dart';

class ImpactGraphView extends StatefulWidget {
  final String rootId;
  final Map<String, dynamic> data;
  const ImpactGraphView(this.rootId, this.data, {super.key});
  @override
  ImpactGraphViewState createState() => ImpactGraphViewState();
}

class ImpactGraphViewState extends State<ImpactGraphView> {
  bool loaded = false;
  Map<String, String> idCategory = {};

  final Graph graph = Graph();

  SugiyamaConfiguration builder = SugiyamaConfiguration()
    ..bendPointShape = CurvedBendPointShape(curveLength: 10);

  @override
  void initState() {
    super.initState();

    builder
      ..nodeSeparation = (25)
      ..levelSeparation = (35)
      ..orientation = SugiyamaConfiguration.ORIENTATION_TOP_BOTTOM;
  }

  @override
  Widget build(BuildContext context) {
    graph.addNode(Node.Id(widget.rootId));
    addToGraph(widget.data["direct"]);
    addToGraph(widget.data["indirect"]);
    addIndirectRelationsToGraph(widget.data["relations"]);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Container(
          constraints: const BoxConstraints(maxHeight: 400, maxWidth: 1000),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.lightBlue.shade100),
            borderRadius: const BorderRadius.all(Radius.circular(15.0)),
          ),
          child: InteractiveViewer(
            alignment: Alignment.center,
            boundaryMargin: const EdgeInsets.all(double.infinity),
            minScale: 0.0001,
            maxScale: 10.6,
            child: OverflowBox(
              minWidth: 0.0,
              minHeight: 0.0,
              maxWidth: double.infinity,
              maxHeight: double.infinity,
              child: GraphView(
                graph: graph,
                algorithm: SugiyamaAlgorithm(builder),
                paint: Paint()
                  ..color = Colors.blue
                  ..strokeWidth = 1
                  ..style = PaintingStyle.stroke,
                builder: (Node node) {
                  final a = node.key!.value as String?;
                  return rectangleWidget(a!);
                },
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget rectangleWidget(String a) {
    return Tooltip(
      message: a,
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(4),
          boxShadow: [
            BoxShadow(
              color: idCategory[a] == "virtual_obj"
                  ? Colors.purple[100]!
                  : Colors.blue[100]!,
              spreadRadius: 1,
            ),
          ],
        ),
        child: Text(a.split(".").last),
      ),
    );
  }

  addToGraph(Map<String, dynamic> value) {
    for (var key in value.keys) {
      var node = Node.Id(key);
      if (!graph.contains(node: node)) {
        graph.addNode(node);
        while (key.contains(".") &&
            key != widget.rootId &&
            !graph.hasPredecessor(node)) {
          final predecessorId = key.substring(0, key.lastIndexOf("."));
          final predecessor = Node.Id(predecessorId);
          graph.addEdge(predecessor, node);
          node = predecessor;
          key = predecessorId;
        }
      }
    }
  }

  addIndirectRelationsToGraph(Map<String, dynamic> value) {
    for (final key in value.keys) {
      final node = Node.Id(key);

      if (!graph.contains(node: node)) {
        graph.addNode(node);
      }
      for (final childId in value[key]) {
        graph.addEdge(
          Node.Id(childId),
          node,
          paint: Paint()..color = Colors.purple,
        );
      }
    }
  }
}
