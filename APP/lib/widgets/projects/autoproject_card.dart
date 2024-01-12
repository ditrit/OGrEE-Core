import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/widgets/projects/project_popup.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';

class AutoProjectCard extends StatelessWidget {
  // final Project project;
  final Namespace namespace;
  final String userEmail;
  final Function parentCallback;
  const AutoProjectCard(
      {Key? key,
      // required this.project,
      required this.namespace,
      required this.userEmail,
      required this.parentCallback})
      : super(key: key);
  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    modifyProjectCallback(String userInput, Project project, bool isCreate,
        Function? parentCallback) async {
      if (userInput == project.name) {
        Navigator.pop(context);
      } else {
        project.name = userInput;
        final messenger = ScaffoldMessenger.of(context);
        var result = await modifyProject(project);
        switch (result) {
          case Success():
            parentCallback!();
            if (context.mounted) Navigator.pop(context);
          case Failure(exception: final exception):
            showSnackBar(messenger, exception.toString(), isError: true);
        }
      }
    }

    deleteProjectCallback(String projectId, Function? parentCallback) async {
      final messenger = ScaffoldMessenger.of(context);
      var result = await deleteProject(projectId);
      switch (result) {
        case Success():
          parentCallback!();
          if (context.mounted) Navigator.pop(context);
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }

    return SizedBox(
      width: 265,
      height: 250,
      child: Card(
        elevation: 3,
        surfaceTintColor: Colors.white,
        margin: const EdgeInsets.all(10),
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  // SizedBox(
                  //   width: 170,
                  //   child: Text(project.name,
                  //       overflow: TextOverflow.clip,
                  //       style: const TextStyle(
                  //           fontWeight: FontWeight.bold, fontSize: 16)),
                  // ),
                  SizedBox(
                    height: 30,
                    child: Badge(
                      backgroundColor: Colors.deepPurple.shade50,
                      label: Text(
                        " ${namespace.name} ",
                        style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Colors.deepPurple.shade900),
                      ),
                    ),
                  ),
                  // CircleAvatar(
                  //   radius: 13,
                  //   child: IconButton(
                  //       splashRadius: 18,
                  //       iconSize: 13,
                  //       padding: const EdgeInsets.all(2),
                  //       onPressed: () => showProjectDialog(
                  //           context,
                  //           project,
                  //           localeMsg.editProject,
                  //           deleteCallback: deleteProjectCallback,
                  //           modifyProjectCallback,
                  //           parentCallback: parentCallback),
                  //       icon: const Icon(
                  //         Icons.mode_edit_outline_rounded,
                  //       )),
                  // )
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text(localeMsg.author),
                  ),
                  Text(
                    "Automatically created",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text("Description :"),
                  ),
                  Text(
                    "View all objects of ${namespace.name} namespace",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: TextButton.icon(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => SelectPage(
                            project: Project(
                                "Physical",
                                "",
                                Namespace.Physical.name,
                                "Automatically generated",
                                "",
                                false,
                                false,
                                false, [], [], []),
                            userEmail: userEmail,
                          ),
                        ),
                      );
                    },
                    icon: const Icon(Icons.play_circle),
                    label: Text(localeMsg.launch)),
              )
            ],
          ),
        ),
      ),
    );
  }
}
