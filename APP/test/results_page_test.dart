// ignore_for_file: prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

import 'common.dart';

void main() {
  testWidgets('ResultPage loads objects', (tester) async {
    await tester.pumpWidget(
      LocalizationsInjApp(
        child: ResultsPage(
          dateRange: "",
          selectedAttrs: [], // should not be const
          selectedObjects: kDataSample["sitePI"]!,
          namespace: Namespace.Test.name,
        ),
      ),
    );
    expect(find.text("Objects"), findsOneWidget);
    for (final obj in kDataSample["sitePI"]!) {
      expect(find.text(obj), findsOneWidget);
    }
  });

  testWidgets('ResultPage adds attributes upon selection', (tester) async {
    await tester.pumpWidget(
      LocalizationsInjApp(
        child: ResultsPage(
          dateRange: "",
          selectedAttrs: [],
          selectedObjects: kDataSample["siteNO"]!,
          namespace: Namespace.Test.name,
        ),
      ),
    );

    await tester.tap(find.byIcon(Icons.add).last);
    await tester.pumpAndSettle();

    for (final attr in ["weight", "vendor"]) {
      await tester.tap(find.textContaining(attr));
      await tester.pumpAndSettle();
      expect(find.text(attr), findsNWidgets(2));
    }
    expect(find.text("45.5"), findsWidgets); //weight
    expect(find.text("test"), findsWidgets); //vendor
  });
}
