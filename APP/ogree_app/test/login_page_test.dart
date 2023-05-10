import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ogree_app/main.dart';

import 'common.dart';

void main() {
  testWidgets('MyApp loads Login Page', (tester) async {
    await tester.pumpWidget(const MyApp());
    final msgs = await getFrenchMessages();
    expect(find.textContaining(msgs["welcome"]!), findsOneWidget);
    expect(find.text('EN'), findsOneWidget);
    expect(find.textContaining('mail'), findsOneWidget);
    expect(find.textContaining(msgs["password"]!), findsNWidgets(2));
    expect(find.textContaining(msgs["login"]!), findsOneWidget);
  });

  testWidgets('Login Page notifies error if email and/or password empty',
      (tester) async {
    await tester.pumpWidget(const MyApp());

    final loginButton = find.textContaining('Se connecter');
    final emailInput = find.ancestor(
        of: find.textContaining('mail'), matching: find.byType(TextFormField));
    final passwordInput = find.ancestor(
        of: find.text('Mot de passe'), matching: find.byType(TextFormField));
    await tester.ensureVisible(loginButton);
    await tester.pumpAndSettle();

    await tester.tap(loginButton);
    await tester.pumpAndSettle();

    expect(find.textContaining('Champ Obligatoire'), findsNWidgets(2));

    await tester.enterText(emailInput, "user@email.com");
    await tester.ensureVisible(loginButton);
    await tester.pumpAndSettle();
    await tester.tap(loginButton);
    await tester.pumpAndSettle();

    expect(find.textContaining('Champ Obligatoire'), findsOneWidget);

    await tester.enterText(emailInput, "");
    await tester.enterText(passwordInput, "password");
    await tester.tap(loginButton);
    await tester.pumpAndSettle();

    expect(find.textContaining('Champ Obligatoire'), findsOneWidget);

    await tester.enterText(emailInput, "user@email.com");
    await tester.enterText(passwordInput, "password");
    await tester.tap(loginButton);
    await tester.pumpAndSettle();

    expect(find.textContaining('Champ Obligatoire'), findsNothing);
  });
}
