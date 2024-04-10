import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/main.dart';
import 'package:ogree_app/widgets/common/language_toggle.dart';

void main() {
  testWidgets('MyApp can languague toogle FR/EN', (tester) async {
    await dotenv.load(fileName: "assets/custom/.env");
    // Create the widget by telling the tester to build it.
    await tester.pumpWidget(const MyApp());

    var titleFinder = find.textContaining('Bienvenue');
    final languageToogleFinder = find.bySubtype<LanguageToggle>();

    expect(titleFinder, findsOneWidget);
    expect(languageToogleFinder, findsOneWidget);

    await tester.tap(languageToogleFinder);
    await tester.pumpAndSettle();

    titleFinder = find.textContaining('Français');
    expect(titleFinder, findsOneWidget);
    titleFinder = find.textContaining('Español');
    expect(titleFinder, findsOneWidget);
    titleFinder = find.textContaining('Português');
    expect(titleFinder, findsOneWidget);
    titleFinder = find.textContaining('English');
    expect(titleFinder, findsOneWidget);

    await tester.tap(titleFinder);
    await tester.pumpAndSettle();

    titleFinder = find.textContaining('Welcome');
    expect(titleFinder, findsOneWidget);
    titleFinder = find.textContaining('Bienvenue');
    expect(titleFinder, findsNothing);
  });
}
