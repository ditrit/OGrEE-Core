import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/widgets/select_date.dart';

import 'common.dart';

void main() {
  testWidgets('SelectDate can toogle between tabs', (tester) async {
    await tester.pumpWidget(const LocalizationsInjApp(child: SelectDate()));
    expect(find.text('Choisir les dates'), findsOneWidget);
    expect(find.textContaining('disponibles'), findsWidgets);

    await tester.tap(find.textContaining("Choisir les dates"));
    await tester.pumpAndSettle();
    expect(find.textContaining('disponibles'), findsNothing);

    await tester.ensureVisible(find.textContaining("Toutes"));
    await tester.pumpAndSettle();
    await tester.tap(find.textContaining("Toutes"));
    await tester.pumpAndSettle();
    expect(find.textContaining('disponibles'), findsWidgets);
  });
}
