import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/select_page.dart';

class ProjectsPage extends StatefulWidget {
  ProjectsPage({super.key});

  @override
  State<ProjectsPage> createState() => _ProjectsPageState();
}

class _ProjectsPageState extends State<ProjectsPage> {
  List<Project>? _projects;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: myAppBar(),
      body: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 80.0, vertical: 20),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Mes projets',
                    style: Theme.of(context).textTheme.headlineLarge),
                Padding(
                  padding: const EdgeInsets.only(right: 10.0, bottom: 10),
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => SelectPage(),
                        ),
                      );
                    },
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: const [
                        Padding(
                          padding: EdgeInsets.symmetric(vertical: 10.0),
                          child: Icon(Icons.add_to_photos),
                        ),
                        Text(
                          "   Créer un nouveau projet",
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
            FutureBuilder(
                future: getProjectData(),
                builder: (context, _) {
                  if (_projects == null) {
                    return const Center(child: CircularProgressIndicator());
                  } else if (_projects!.isNotEmpty) {
                    return Expanded(
                      child: GridView.extent(
                        padding: const EdgeInsets.only(top: 5),
                        maxCrossAxisExtent: 270,
                        children: getProjectCards(context),
                      ),
                    );
                  } else {
                    // Empty messages
                    return Text("Unable to get projects");
                  }
                }),
          ],
        ),
      ),
    );
  }

  getProjectData() async {
    _projects = await fetchProjects();
  }

  getProjectCards(context) {
    List<Card> cards = [];
    for (var project in _projects!) {
      cards.add(Card(
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
                    width: 174,
                    child: Text("Projet ${project.name}",
                        overflow: TextOverflow.clip,
                        style: TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16)),
                  ),
                  CircleAvatar(
                    radius: 13,
                    backgroundColor: Colors.blue,
                    child: IconButton(
                        splashRadius: 18,
                        iconSize: 13,
                        padding: EdgeInsets.all(2),
                        onPressed: () => showCustomDialog(
                            context,
                            project,
                            "Editer ce projet",
                            "Supprimer",
                            Icons.delete,
                            deleteProjectCallback,
                            modifyProjectCallback),
                        icon: Icon(
                          Icons.mode_edit_outline_rounded,
                          color: Colors.white,
                        )),
                  )
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Auteur :"),
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
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Dernière modification :"),
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
                          ),
                        ),
                      );
                    },
                    icon: const Icon(Icons.play_circle),
                    label: const Text("Lancer")),
              )
            ],
          ),
        ),
      ));
    }
    return cards;
  }

  modifyProjectCallback(
      String userInput, Project project, bool isCreate) async {
    if (userInput == project.name) {
      Navigator.pop(context);
    } else {
      project.name = userInput;
      var response = await modifyProject(project);
      if (response == "") {
        setState(() {});
        Navigator.pop(context);
      } else {
        showSnackBar(context, response, isError: true);
      }
    }
  }

  deleteProjectCallback(String projectId) async {
    var response = await deleteProject(projectId);
    if (response == "") {
      setState(() {});
      Navigator.pop(context);
    } else {
      showSnackBar(context, response, isError: true);
    }
  }
}
