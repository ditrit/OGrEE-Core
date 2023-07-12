import 'package:flutter/material.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import '../../common/theme.dart';

void showProjectDialog(
  BuildContext context,
  Project project,
  String title,
  Function saveCallback, {
  bool isCreate = false,
  Function? parentCallback,
  Function? deleteCallback,
}) {
  String editInput = project.name;
  showGeneralDialog(
    context: context,
    barrierLabel: "Barrier",
    barrierDismissible: true,
    barrierColor: Colors.black.withOpacity(0.1),
    transitionDuration: const Duration(milliseconds: 700),
    pageBuilder: (context, _, __) {
      final localeMsg = AppLocalizations.of(context)!;
      return Center(
        child: Container(
          height: 200,
          width: 500,
          margin: const EdgeInsets.symmetric(horizontal: 20),
          decoration: PopupDecoration,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 8),
            child: Material(
              color: Colors.white,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(title,
                      style: Theme.of(context).textTheme.headlineMedium),
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 25),
                    child: TextFormField(
                      initialValue: project.name,
                      onChanged: (value) => editInput = value,
                      decoration: GetFormInputDecoration(
                          false, localeMsg.projectName,
                          icon: Icons.edit_outlined),
                      cursorWidth: 1.3,
                      style: const TextStyle(fontSize: 14),
                    ),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton.icon(
                        style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.blue.shade900),
                        onPressed: () => Navigator.pop(context),
                        label: Text(localeMsg.cancel),
                        icon: const Icon(
                          Icons.cancel_outlined,
                          size: 16,
                        ),
                      ),
                      deleteCallback != null
                          ? TextButton.icon(
                              style: OutlinedButton.styleFrom(
                                  foregroundColor: Colors.red.shade900),
                              onPressed: () =>
                                  deleteCallback(project.id, parentCallback),
                              label: Text(localeMsg.delete),
                              icon: const Icon(
                                Icons.delete,
                                size: 16,
                              ),
                            )
                          : Container(),
                      const SizedBox(width: 10),
                      ElevatedButton(
                        onPressed: () async {
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
