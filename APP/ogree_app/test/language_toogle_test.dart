import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/main.dart';

void main() {
  testWidgets('MyApp can languague toogle FR/EN', (tester) async {
    await dotenv.load(fileName: "assets/custom/.env");
    // Create the widget by telling the tester to build it.
    await tester.pumpWidget(const MyApp());

    var titleFinder = find.textContaining('Bienvenue');
    final languageToogleFinder = find.text('EN');

    expect(titleFinder, findsOneWidget);
    expect(languageToogleFinder, findsOneWidget);

    await tester.tap(languageToogleFinder);
    await tester.pumpAndSettle();

    titleFinder = find.textContaining('Welcome');
    expect(titleFinder, findsOneWidget);
  });
}
