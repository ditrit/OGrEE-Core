import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/project.dart';

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
      final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
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
                      style: Theme.of(context).textTheme.headlineMedium,),
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 25),
                    child: TextFormField(
                      initialValue: project.name,
                      onChanged: (value) => editInput = value,
                      decoration: GetFormInputDecoration(
                          false, localeMsg.projectName,
                          icon: Icons.edit_outlined,),
                      cursorWidth: 1.3,
                      style: const TextStyle(fontSize: 14),
                    ),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton.icon(
                        style: OutlinedButton.styleFrom(
                            padding: isSmallDisplay
                                ? const EdgeInsets.symmetric(horizontal: 8)
                                : null,
                            foregroundColor: Colors.blue.shade900,),
                        onPressed: () => Navigator.pop(context),
                        label: Text(localeMsg.cancel),
                        icon: const Icon(
                          Icons.cancel_outlined,
                          size: 16,
                        ),
                      ),
                      if (deleteCallback != null) isSmallDisplay
                              ? IconButton(
                                  iconSize: 16,
                                  onPressed: () => deleteCallback(
                                      project.id, parentCallback,),
                                  icon: Icon(
                                    Icons.delete,
                                    color: Colors.red.shade900,
                                  ),)
                              : TextButton.icon(
                                  style: OutlinedButton.styleFrom(
                                      foregroundColor: Colors.red.shade900,),
                                  onPressed: () => deleteCallback(
                                      project.id, parentCallback,),
                                  label: Text(
                                      isSmallDisplay ? "" : localeMsg.delete,),
                                  icon: const Icon(
                                    Icons.delete,
                                    size: 16,
                                  ),
                                ) else Container(),
                      SizedBox(width: isSmallDisplay ? 0 : 10),
                      ElevatedButton(
                        onPressed: () async {
                          if (editInput == "") {
                            showSnackBar(ScaffoldMessenger.of(context),
                                localeMsg.mandatoryProjectName,
                                isError: true,);
                          } else {
                            saveCallback(
                                editInput, project, isCreate, parentCallback,);
                          }
                        },
                        style: isSmallDisplay
                            ? ElevatedButton.styleFrom(
                                padding:
                                    const EdgeInsets.symmetric(horizontal: 8),
                              )
                            : null,
                        child: Text(localeMsg.save),
                      ),
                    ],
                  ),
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
