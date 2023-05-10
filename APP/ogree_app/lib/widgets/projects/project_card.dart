import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/widgets/projects/project_popup.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/select_page.dart';

class ProjectCard extends StatelessWidget {
  final Project project;
  final String userEmail;
  final Function parentCallback;
  const ProjectCard(
      {Key? key,
      required this.project,
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
        var response = await modifyProject(project);
        if (response == "") {
          parentCallback!();
          Navigator.pop(context);
        } else {
          showSnackBar(context, response, isError: true);
        }
      }
    }

    deleteProjectCallback(String projectId, Function? parentCallback) async {
      var response = await deleteProject(projectId);
      if (response == "") {
        parentCallback!();
        Navigator.pop(context);
      } else {
        showSnackBar(context, response, isError: true);
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
                  Container(
                    width: 170,
                    child: Text(project.name,
                        overflow: TextOverflow.clip,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16)),
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
                            localeMsg.delete,
                            Icons.delete,
                            deleteProjectCallback,
                            modifyProjectCallback,
                            parentCallback: parentCallback),
                        icon: const Icon(
                          Icons.mode_edit_outline_rounded,
                        )),
                  )
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
                    " ${project.authorLastUpdate}",
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
                    " ${project.lastUpdate}",
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
                    label: Text(localeMsg.launch)),
              )
            ],
          ),
        ),
      ),
    );
  }
}
