import 'package:flutter/material.dart';
import 'package:ogree_app/widgets/language_toggle.dart';
import 'package:ogree_app/widgets/login_card.dart';
import 'package:ogree_app/widgets/reset_card.dart';

class LoginPage extends StatefulWidget {
  final bool isPasswordReset;
  final String resetToken;
  const LoginPage(
      {super.key, this.isPasswordReset = false, this.resetToken = ""});
  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          image: DecorationImage(
            image: AssetImage("assets/server_background.png"),
            fit: BoxFit.cover,
          ),
        ),
        child: CustomScrollView(slivers: [
          SliverFillRemaining(
            hasScrollBody: false,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Align(
                  alignment: Alignment.topCenter,
                  child: LanguageToggle(),
                ),
                const SizedBox(height: 5),
                widget.isPasswordReset
                    ? ResetCard(
                        token: widget.resetToken,
                      )
                    : const LoginCard(),
              ],
            ),
          )
        ]),
      ),
    );
  }
}
