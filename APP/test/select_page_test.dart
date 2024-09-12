import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/pages/select_page.dart';

import 'common.dart';

void main() {
  testWidgets('SelectPage can select date and namespace', (tester) async {
    await tester.pumpWidget(LocalizationsInjApp(
        child: SelectPage(
      userEmail: 'user@test.com',
    ),),);

    // Date
    expect(find.text('Choisir les dates'), findsOneWidget);
    // var defaultDate = DateFormat('dd/MM/yyyy').format(DateTime.now());
    // expect(find.text(defaultDate), findsOneWidget);

    // Next
    await tester.ensureVisible(find.text("Suivant"));
    await tester.pumpAndSettle();
    await tester.tap(find.text("Suivant").first);
    await tester.pumpAndSettle();

    // Namespace
    expect(find.text("Physical"), findsNWidgets(1));
    expect((tester.widget(find.text("Physical")) as Text).style!.color,
        Colors.blue,);
    expect((tester.widget(find.text("Logical")) as Text).style!.color, null);
    // expect(find.text(defaultDate, skipOffstage: false), findsOneWidget);
    // await tester.tap(find.text("Logical"));
    // await tester.pumpAndSettle();
    // expect((tester.widget(find.text("Logical")) as Text).style!.color,
    //     Colors.blue);
    // expect((tester.widget(find.text("Physical").at(1)) as Text).style!.color,
    //     Colors.black);
  });
}
