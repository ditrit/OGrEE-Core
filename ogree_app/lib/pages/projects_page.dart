import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/tenants/popups/create_tenant_popup.dart';
import 'package:ogree_app/widgets/projects/project_card.dart';
import 'package:ogree_app/widgets/tenants/tenant_card.dart';

class ProjectsPage extends StatefulWidget {
  final String userEmail;
  final bool isTenantMode;
  ProjectsPage(
      {super.key, required this.userEmail, required this.isTenantMode});

  @override
  State<ProjectsPage> createState() => _ProjectsPageState();
}

class _ProjectsPageState extends State<ProjectsPage> {
  List<Project>? _projects;
  List<Tenant>? _tenants;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: myAppBar(context, widget.userEmail,
          isTenantMode: widget.isTenantMode),
      body: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 80.0, vertical: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(widget.isTenantMode ? "Tenants" : localeMsg.myprojects,
                    style: Theme.of(context).textTheme.headlineLarge),
                Padding(
                  padding: const EdgeInsets.only(right: 10.0, bottom: 10),
                  child: createProjectButton(),
                ),
              ],
            ),
            FutureBuilder(
                future: getProjectData(),
                builder: (context, _) {
                  if (_projects == null && _tenants == null) {
                    return const Center(child: CircularProgressIndicator());
                  } else if (_projects != null && _projects!.isNotEmpty) {
                    return Expanded(
                      child: SingleChildScrollView(
                        child: Wrap(
                          spacing: 5,
                          children: getCards(context),
                        ),
                      ),
                    );
                  } else if (_tenants != null && _tenants!.isNotEmpty) {
                    return Expanded(
                      child: SingleChildScrollView(
                        child: Wrap(
                          spacing: 5,
                          children: getCards(context),
                        ),
                      ),
                    );
                  } else {
                    // Empty messages
                    return Text(localeMsg.noProjects);
                  }
                }),
          ],
        ),
      ),
    );
  }

  refreshFromChildren() {
    setState(() {});
  }

  getProjectData() async {
    if (widget.isTenantMode) {
      _tenants = await fetchTenants();
    } else {
      _projects = await fetchProjects(widget.userEmail);
    }
  }

  createProjectButton() {
    final localeMsg = AppLocalizations.of(context)!;
    return ElevatedButton(
      onPressed: () {
        if (widget.isTenantMode) {
          showCustomPopup(
              context, CreateTenantPopup(parentCallback: refreshFromChildren));
        } else {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => SelectPage(userEmail: widget.userEmail),
            ),
          );
        }
      },
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 10.0),
            child: Icon(Icons.add_to_photos),
          ),
          Text(
              widget.isTenantMode ? localeMsg.newTenant : localeMsg.newProject),
        ],
      ),
    );
  }

  getCards(context) {
    List<Widget> cards = [];
    if (widget.isTenantMode) {
      for (var tenant in _tenants!) {
        cards.add(TenantCard(
          tenant: tenant,
          parentCallback: refreshFromChildren,
        ));
      }
    } else {
      for (var project in _projects!) {
        cards.add(ProjectCard(
          project: project,
          userEmail: widget.userEmail,
          parentCallback: refreshFromChildren,
        ));
      }
    }
    return cards;
  }
}
