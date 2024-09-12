import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

import 'common.dart';

void main() {
  testWidgets('SelectObjects expands and collapses tree', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1000, 1000));
    await tester.pumpWidget(LocalizationsInjApp(
        child: SelectObjects(
      dateRange: "",
      load: true,
      namespace: Namespace.Test,
    ),),);

    final expandButton = find.text("Développer tout");
    await tester.tap(expandButton);
    await tester.pumpAndSettle();
    for (final node in kDataSample["sitePI.B1.1.rack1"]!) {
      expect(find.text(node.substring(node.lastIndexOf(".") + 1)),
          findsAtLeastNWidgets(1),);
    }

    final collapseButton = find.text("Réduire tout");
    await tester.tap(collapseButton);
    await tester.pumpAndSettle();
    for (final node in kDataSample["sitePI.B1.1.rack1"]!) {
      expect(
          find.text(node.substring(node.lastIndexOf(".") + 1)), findsNothing,);
    }
  });

  testWidgets('SelectObjects toogles tree selection', (tester) async {
    await tester.pumpWidget(LocalizationsInjApp(
        child: SelectObjects(
      dateRange: "",
      load: true,
      namespace: Namespace.Test,
    ),),);

    final expandButton = find.text("Sélectionner tout");
    await tester.tap(expandButton);
    await tester.pumpAndSettle();
    expect(find.textContaining("site"), findsNWidgets(23));

    // Scroll until deselect button appears and tap it
    final collapseButton = find.text("Désélectionner tout");
    await tester.scrollUntilVisible(
      collapseButton,
      500.0,
      scrollable: find.ancestor(
          of: find.textContaining("siteNO.BB1"),
          matching: find.byType(Scrollable),),
    );
    await tester.tap(collapseButton);
    await tester.pumpAndSettle();
    expect(find.textContaining("site"), findsNWidgets(4));
  });

  testWidgets('SelectObjects can find an object', (tester) async {
    await tester.pumpWidget(LocalizationsInjApp(
        child: Scaffold(
      body: SelectObjects(
        dateRange: "",
        load: true,
        namespace: Namespace.Test,
      ),
    ),),);

    const searchStr = "rack2.devB.devB-2";
    final searchInput =
        find.ancestor(of: find.text('ID'), matching: find.byType(TextField));
    await tester.enterText(searchInput, searchStr);
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();

    final List<String> parents = searchStr.split(".");
    parents.removeAt(0);
    for (final name in parents) {
      expect(find.text(name), findsOneWidget);
    }
    expect(find.text("devC-2"), findsNothing);
  });

  testWidgets('SelectObjects can filter objects', (tester) async {
    return;
    await tester.binding.setSurfaceSize(const Size(1000, 1000));
    await tester.pumpWidget(LocalizationsInjApp(
        child: SelectObjects(
      dateRange: "",
      load: true,
      namespace: Namespace.Test,
    ),),);

    // all data is there
    for (final site in kDataSample[kRootId]!) {
      expect(find.text(site), findsOneWidget);
    }

    // select building input and get suggestions
    final searchBldg = find.text("Building");
    await tester.press(searchBldg, warnIfMissed: false);
    await tester.pumpAndSettle();
    for (final building in ['sitePA.A1', 'sitePA.A2', 'sitePI.B1']) {
      expect(find.text(building), findsOneWidget);
    }

    // type in a suggestion filter
    var searchInput =
        find.ancestor(of: searchBldg, matching: find.byType(TextFormField));
    await tester.enterText(searchInput, "sitePI.B1");
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();

    // check if tree is filtered
    for (final site in kDataSample[kRootId]!) {
      expect(find.text(site), findsNothing);
    }
    final expandButton = find.text("Développer tout");
    await tester.tap(expandButton);
    await tester.pumpAndSettle();
    expect(find.text("sitePI.B1"), findsNWidgets(3));
    for (final obj in ["rack1", "rack2"]) {
      expect(find.text(obj), findsOneWidget);
    }
    for (final obj in ["devA", "devB", "devC", "devD"]) {
      expect(find.text(obj), findsNWidgets(2));
    }

    // add a second filter
    final searchRack = find.text("Rack");
    searchInput =
        find.ancestor(of: searchRack, matching: find.byType(TextFormField));
    await tester.enterText(searchInput, "sitePI.B1.1.rack2");
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();

    // check if tree is filtered
    for (final site in kDataSample[kRootId]!) {
      expect(find.text(site), findsNothing);
    }
    await tester.ensureVisible(expandButton);
    await tester.pumpAndSettle();
    await tester.tap(expandButton);
    await tester.pumpAndSettle();
    expect(find.text("sitePI.B1"), findsNWidgets(3));
    for (final obj in ["devA", "devB", "devC", "devD"]) {
      expect(find.text(obj), findsOneWidget);
    }
    for (final obj in [
      "rack1",
      "B2",
      "NO",
      "PA",
      "PB",
    ]) {
      expect(find.textContaining(obj), findsNothing);
    }

    // clear all filters
    final clearButton = find.text("Effacer tout");
    await tester.tap(clearButton);
    await tester.pumpAndSettle();
    for (final obj in ["devA", "devB", "devC", "devD"]) {
      expect(find.text(obj), findsNothing);
    }
    for (final obj in kDataSample[kRootId]!) {
      expect(find.text(obj), findsOneWidget);
    }

    // resets the screen to its original size after the test end
    tester.view.resetPhysicalSize();
  });
}
