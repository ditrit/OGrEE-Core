import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

void showProjectDialog(
    BuildContext context,
    Project project,
    String title,
    String cancelBtnTitle,
    IconData cancelIcon,
    Function cancelCallback,
    Function saveCallback,
    {bool isCreate = false,
    Function? parentCallback}) {
  String editInput = project.name;
  const inputStyle = OutlineInputBorder(
    borderSide: BorderSide(
      color: Colors.grey,
      width: 1,
    ),
  );
  showGeneralDialog(
    context: context,
    barrierLabel: "Barrier",
    barrierDismissible: true,
    barrierColor: Colors.black.withOpacity(0.5),
    transitionDuration: const Duration(milliseconds: 700),
    pageBuilder: (context, _, __) {
      final localeMsg = AppLocalizations.of(context)!;
      return Center(
        child: Container(
          height: 240,
          width: 500,
          margin: const EdgeInsets.symmetric(horizontal: 20),
          decoration: BoxDecoration(
              color: Colors.white, borderRadius: BorderRadius.circular(40)),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 10),
            child: Material(
              color: Colors.white,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(title, style: Theme.of(context).textTheme.headlineLarge),
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 40),
                    child: TextFormField(
                      initialValue: project.name,
                      onChanged: (value) => editInput = value,
                      decoration: InputDecoration(
                        labelText: localeMsg.projectName,
                        labelStyle: GoogleFonts.inter(
                          fontSize: 12,
                          color: Colors.black,
                        ),
                        enabledBorder: inputStyle,
                        focusedBorder: inputStyle,
                      ),
                    ),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton.icon(
                        style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.red.shade900),
                        onPressed: () =>
                            cancelCallback(project.id, parentCallback),
                        label: Text(cancelBtnTitle),
                        icon: Icon(
                          cancelIcon,
                          size: 16,
                        ),
                      ),
                      const SizedBox(width: 15),
                      ElevatedButton(
                        onPressed: () async {
                          print(editInput);
                          if (editInput == "") {
                            showSnackBar(
                                context, localeMsg.mandatoryProjectName,
                                isError: true);
                          } else {
                            saveCallback(
                                editInput, project, isCreate, parentCallback);
                          }
                        },
                        child: Text(localeMsg.save),
                      )
                    ],
                  )
                ],
              ),
            ),
          ),
        ),
      );
    },
    transitionBuilder: (_, anim, __, child) {
      Tween<Offset> tween;
      if (anim.status == AnimationStatus.reverse) {
        tween = Tween(begin: const Offset(-1, 0), end: Offset.zero);
      } else {
        tween = Tween(begin: const Offset(1, 0), end: Offset.zero);
      }
      return SlideTransition(
        position: tween.animate(anim),
        child: FadeTransition(
          opacity: anim,
          child: child,
        ),
      );
    },
  );
}
