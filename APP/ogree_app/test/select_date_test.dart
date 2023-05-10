import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/widgets/select_date.dart';

import 'common.dart';

void main() {
  testWidgets('SelectDate can toogle between tabs', (tester) async {
    await tester.pumpWidget(LocalizationsInjApp(child: SelectDate()));
    expect(find.text('Choisir les dates'), findsOneWidget);

    await tester.tap(find.textContaining("dernier"));
    await tester.pumpAndSettle();
    expect(find.textContaining('Données mises à jour le'), findsOneWidget);

    await tester.ensureVisible(find.textContaining("enregistré"));
    await tester.pumpAndSettle();
    await tester.tap(find.textContaining("enregistré"));
    await tester.pumpAndSettle();
    expect(find.textContaining('Jeu'), findsWidgets);
  });
}
