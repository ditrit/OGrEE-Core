import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:graphview/GraphView.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';

class ObjectGraphView extends StatefulWidget {
  final String rootId;
  const ObjectGraphView(this.rootId, {super.key});
  @override
  ObjectGraphViewState createState() => ObjectGraphViewState();
}

class ObjectGraphViewState extends State<ObjectGraphView> {
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
    return FutureBuilder(
      future: !loaded ? getObject() : null,
      builder: (context, _) {
        if (!loaded) {
          return const Center(child: CircularProgressIndicator());
        }
        return Center(
          child: Container(
            constraints: const BoxConstraints(maxHeight: 520, maxWidth: 800),
            margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 20),
            decoration: PopupDecoration,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
              child: Scaffold(
                backgroundColor: Colors.white,
                body: Column(
                  children: [
                    Center(
                      child: Text(
                        AppLocalizations.of(context)!.viewGraph,
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                    ),
                    Expanded(
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
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        ElevatedButton.icon(
                          onPressed: () {
                            Navigator.of(context).pop();
                          },
                          label: const Text("OK"),
                          icon: const Icon(Icons.thumb_up, size: 16),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget rectangleWidget(String a) {
    return Tooltip(
      message: a,
      child: Container(
        padding: const EdgeInsets.all(16),
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

  Future<void> getObject() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchObjectChildren(widget.rootId);
    switch (result) {
      case Success(value: final value):
        addToGraph(value);
        loaded = true;
        return;
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }

  Node addToGraph(Map<String, dynamic> value) {
    final node = Node.Id(value["id"]);
    graph.addNode(node);
    idCategory[value["id"]] = value["category"];
    if (value["attributes"] != null && value["attributes"]["vlinks"] != null) {
      for (final vlink in List<String>.from(value["attributes"]["vlinks"])) {
        graph.addEdge(
          node,
          Node.Id(vlink),
          paint: Paint()..color = Colors.purple,
        );
      }
    }
    if (value["children"] != null) {
      for (final child in List<Map<String, dynamic>>.from(value["children"])) {
        final childNode = addToGraph(child);
        graph.addEdge(node, childNode);
      }
    }
    return node;
  }
}
