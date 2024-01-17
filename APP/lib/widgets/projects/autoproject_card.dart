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
    var color = Colors.teal;
    if (namespace == Namespace.Logical) {
      color = Colors.deepPurple;
    } else if (namespace == Namespace.Organisational) {
      color = Colors.deepOrange;
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
              SizedBox(
                height: 30,
                child: Badge(
                  backgroundColor: color.shade50,
                  label: Text(
                    " ${namespace.name} ",
                    style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: color.shade900),
                  ),
                ),
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
                    "View all objects of namespace ${namespace.name}",
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
                                namespace.name,
                                "",
                                namespace.name,
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
