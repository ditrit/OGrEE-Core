import 'package:flutter/material.dart';
import 'package:ogree_app/widgets/common/language_toggle.dart';
import 'package:ogree_app/widgets/login/login_card.dart';
import 'package:ogree_app/widgets/login/reset_card.dart';

class LoginPage extends StatefulWidget {
  final bool isPasswordReset;
  final String resetToken;
  const LoginPage(
      {super.key, this.isPasswordReset = false, this.resetToken = "",});
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
              children: [
                const Align(
                  alignment: Alignment.topCenter,
                  child: LanguageToggle(),
                ),
                const SizedBox(height: 5),
                if (widget.isPasswordReset) ResetCard(
                        token: widget.resetToken,
                      ) else const LoginCard(),
              ],
            ),
          ),
        ],),
      ),
    );
  }
}
