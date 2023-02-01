import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/pages/projects_page.dart';

class LoginPage extends StatefulWidget {
  static String tag = 'login-page';

  const LoginPage({super.key});
  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  bool _isChecked = false;
  static const inputStyle = OutlineInputBorder(
    borderSide: BorderSide(
      color: Colors.grey,
      width: 1,
    ),
  );

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        image: DecorationImage(
          image: AssetImage("assets/server_background.png"),
          fit: BoxFit.cover,
        ),
      ),
      child: Center(
        child: Card(
          child: Container(
            constraints: const BoxConstraints(maxWidth: 550, maxHeight: 600),
            padding: const EdgeInsets.symmetric(horizontal: 100, vertical: 50),
            child: ListView(
              shrinkWrap: true,
              children: [
                Center(
                    child: Text('Bienvenue sur OGrEE',
                        style: Theme.of(context).textTheme.headlineLarge)),
                const SizedBox(height: 8),
                Center(
                  child: Text(
                    'Connectez-vous à votre espace :',
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                ),
                const SizedBox(height: 32),
                Center(
                  child: Image.asset(
                    "assets/edf_logo.png",
                    height: 30,
                  ),
                ),
                const SizedBox(height: 32),
                TextField(
                  decoration: InputDecoration(
                    labelText: 'E-mail',
                    hintText: 'abc@example.com',
                    labelStyle: GoogleFonts.inter(
                      fontSize: 12,
                      color: Colors.black,
                    ),
                    enabledBorder: inputStyle,
                    focusedBorder: inputStyle,
                  ),
                ),
                const SizedBox(height: 20),
                TextField(
                  obscureText: true,
                  decoration: InputDecoration(
                    labelText: 'Mot de passe',
                    hintText: '********',
                    labelStyle: GoogleFonts.inter(
                      fontSize: 12,
                      color: Colors.black,
                    ),
                    enabledBorder: inputStyle,
                    focusedBorder: inputStyle,
                  ),
                ),
                const SizedBox(height: 25),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        SizedBox(
                          height: 24,
                          width: 24,
                          child: Checkbox(
                            value: _isChecked,
                            onChanged: (bool? value) =>
                                setState(() => _isChecked = value!),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Rester connecté',
                          style: GoogleFonts.inter(
                            fontSize: 14,
                            color: Colors.black,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(width: 25),
                    Text(
                      'Mot de passe oublié ?',
                      style: GoogleFonts.inter(
                        fontSize: 14,
                        color: const Color.fromARGB(255, 0, 84, 152),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 40),
                Align(
                  child: TextButton(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => ProjectsPage(),
                        ),
                      );
                    },
                    style: TextButton.styleFrom(
                      backgroundColor: Theme.of(context).colorScheme.primary,
                      padding: const EdgeInsets.symmetric(
                        vertical: 20,
                        horizontal: 20,
                      ),
                    ),
                    child: Text(
                      'Se connecter',
                      style: GoogleFonts.inter(
                        fontSize: 14,
                        color: Colors.white,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 15),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
