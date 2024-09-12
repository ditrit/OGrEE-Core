import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/projects/project_popup.dart';

class ProjectCard extends StatelessWidget {
  final Project project;
  final String userEmail;
  final Function parentCallback;
  const ProjectCard({
    super.key,
    required this.project,
    required this.userEmail,
    required this.parentCallback,
  });
  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    modifyProjectCallback(
      String userInput,
      Project project,
      bool isCreate,
      Function? parentCallback,
    ) async {
      if (userInput == project.name) {
        Navigator.pop(context);
      } else {
        project.name = userInput;
        final messenger = ScaffoldMessenger.of(context);
        final result = await modifyProject(project);
        switch (result) {
          case Success():
            parentCallback!();
            Navigator.pop(context);
          case Failure(exception: final exception):
            showSnackBar(messenger, exception.toString(), isError: true);
        }
      }
    }

    deleteProjectCallback(String projectId, Function? parentCallback) async {
      final messenger = ScaffoldMessenger.of(context);
      final result = await deleteObject(projectId, "project");
      switch (result) {
        case Success():
          parentCallback!();
          Navigator.pop(context);
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
                  if (project.isImpact)
                    Padding(
                      padding: const EdgeInsets.only(right: 4),
                      child: Tooltip(
                        message: localeMsg.impactAnalysis,
                        child: const Icon(
                          Icons.settings_suggest,
                          size: 14,
                          color: Colors.black,
                        ),
                      ),
                    )
                  else
                    Container(),
                  SizedBox(
                    width: 160,
                    child: Text(
                      project.name,
                      overflow: TextOverflow.clip,
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  CircleAvatar(
                    radius: 13,
                    child: IconButton(
                      splashRadius: 18,
                      iconSize: 13,
                      padding: const EdgeInsets.all(2),
                      onPressed: () => showProjectDialog(
                        context,
                        project,
                        localeMsg.editProject,
                        deleteCallback: deleteProjectCallback,
                        modifyProjectCallback,
                        parentCallback: parentCallback,
                      ),
                      icon: const Icon(
                        Icons.mode_edit_outline_rounded,
                      ),
                    ),
                  ),
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
                    project.authorLastUpdate,
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text(localeMsg.lastUpdate),
                  ),
                  Text(
                    project.lastUpdate,
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
                          project: project,
                          userEmail: userEmail,
                        ),
                      ),
                    );
                  },
                  icon: const Icon(Icons.play_circle),
                  label: Text(localeMsg.launch),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
