import 'package:flutter/material.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/widgets/projects/autoproject_card.dart';
import 'package:ogree_app/widgets/tenants/popups/create_tenant_popup.dart';
import 'package:ogree_app/widgets/projects/project_card.dart';
import 'package:ogree_app/widgets/tenants/tenant_card.dart';
import 'package:ogree_app/widgets/tools/create_netbox_popup.dart';
import 'package:ogree_app/widgets/tools/create_opendcim_popup.dart';
import 'package:ogree_app/widgets/tools/download_cli_popup.dart';
import 'package:ogree_app/widgets/tools/tool_card.dart';

class ProjectsPage extends StatefulWidget {
  final String userEmail;
  final bool isTenantMode;
  const ProjectsPage(
      {super.key, required this.userEmail, required this.isTenantMode});

  @override
  State<ProjectsPage> createState() => _ProjectsPageState();
}

class _ProjectsPageState extends State<ProjectsPage> {
  List<Project>? _projects;
  List<Tenant>? _tenants;
  List<DockerContainer>? _tools;
  bool _isSmallDisplay = false;
  bool _hasNetbox = false;
  bool _hasOpenDcim = false;
  bool _gotData = false;

  @override
  Widget build(BuildContext context) {
    _isSmallDisplay = MediaQuery.of(context).size.width < 600;
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: myAppBar(context, widget.userEmail,
          isTenantMode: widget.isTenantMode),
      body: Padding(
        padding: EdgeInsets.symmetric(
            horizontal: _isSmallDisplay ? 40 : 80.0, vertical: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                widget.isTenantMode
                    ? Row(
                        children: [
                          Text("Applications",
                              style: Theme.of(context).textTheme.headlineLarge),
                          IconButton(
                              onPressed: () => setState(() {
                                    _gotData = false;
                                  }),
                              icon: const Icon(Icons.refresh))
                        ],
                      )
                    : Text(localeMsg.myprojects,
                        style: Theme.of(context).textTheme.headlineLarge),
                Row(
                  children: [
                    Padding(
                      padding: const EdgeInsets.only(right: 10.0, bottom: 10),
                      child: createProjectButton(),
                    ),
                    widget.isTenantMode
                        ? Padding(
                            padding:
                                const EdgeInsets.only(right: 10.0, bottom: 10),
                            child: createToolsButton(),
                          )
                        : Container(),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 3),
            FutureBuilder(
                future: _gotData ? null : getProjectData(),
                builder: (context, _) {
                  if (!_gotData) {
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
                  } else if ((_tenants != null && _tenants!.isNotEmpty) ||
                      (_tools != null && _tools!.isNotEmpty)) {
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
    setState(() {
      _gotData = false;
    });
  }

  getProjectData() async {
    final messenger = ScaffoldMessenger.of(context);
    if (widget.isTenantMode) {
      final result = await fetchApplications();
      switch (result) {
        case Success(value: final value):
          var (tenants, tools) = value;
          _tenants = tenants;
          for (var tenant in tenants) {
            final result = await fetchTenantDockerInfo(tenant.name);
            switch (result) {
              case Success(value: final value):
                List<DockerContainer> dockerInfo = value;
                if (dockerInfo.isEmpty) {
                  tenant.status = TenantStatus.unavailable;
                } else {
                  int runCount = 0;
                  for (var container in dockerInfo) {
                    if (container.status.contains("run")) {
                      runCount++;
                    }
                  }
                  if (runCount == dockerInfo.length) {
                    tenant.status = TenantStatus.running;
                  } else if (runCount > 0) {
                    tenant.status = TenantStatus.partialRun;
                  } else {
                    tenant.status = TenantStatus.notRunning;
                  }
                }
              case Failure():
                tenant.status = TenantStatus.unavailable;
            }
          }
          _tools = tools;
          setState(() {
            _gotData = true;
          });
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
          _tenants = [];
      }
    } else {
      final result = await fetchProjects(widget.userEmail);
      switch (result) {
        case Success(value: final value):
          _projects = value;
          setState(() {
            _gotData = true;
          });
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
          _projects = [];
      }
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
          Padding(
            padding: EdgeInsets.only(
                top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10),
            child: const Icon(Icons.add_to_photos),
          ),
          _isSmallDisplay
              ? Container()
              : Text(widget.isTenantMode
                  ? "${localeMsg.create} tenant"
                  : localeMsg.newProject),
        ],
      ),
    );
  }

  createToolsButton() {
    final localeMsg = AppLocalizations.of(context)!;
    List<PopupMenuEntry<Tools>> entries = <PopupMenuEntry<Tools>>[
      PopupMenuItem(
        value: Tools.netbox,
        child: Text("${localeMsg.create} Netbox"),
      ),
      PopupMenuItem(
        value: Tools.opendcim,
        child: Text("${localeMsg.create} OpenDCIM"),
      ),
      PopupMenuItem(
        value: Tools.cli,
        child: Text(localeMsg.downloadCli),
      ),
    ];

    return ElevatedButton(
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.green.shade600,
        foregroundColor: Colors.white,
      ),
      onPressed: () {},
      child: PopupMenuButton<Tools>(
        offset: const Offset(20, 40),
        onSelected: (value) {
          switch (value) {
            case Tools.netbox:
              if (_hasNetbox) {
                showSnackBar(ScaffoldMessenger.of(context),
                    localeMsg.onlyOneTool("Netbox"));
              } else {
                showCustomPopup(context,
                    CreateNetboxPopup(parentCallback: refreshFromChildren));
              }
              break;
            case Tools.opendcim:
              if (_hasOpenDcim) {
                showSnackBar(ScaffoldMessenger.of(context),
                    localeMsg.onlyOneTool("OpenDCIM"));
              } else {
                showCustomPopup(context,
                    CreateOpenDcimPopup(parentCallback: refreshFromChildren));
              }
              break;
            case Tools.cli:
              showCustomPopup(context, const DownloadCliPopup());
              break;
          }
        },
        itemBuilder: (_) => entries,
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: EdgeInsets.only(
                  top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10),
              child: const Icon(Icons.timeline),
            ),
            _isSmallDisplay ? Container() : Text(localeMsg.tools),
          ],
        ),
      ),
    );
  }

  getCards(context) {
    List<Widget> cards = [];
    if (widget.isTenantMode) {
      if (_tenants != null && _tenants!.isNotEmpty) {
        for (var tenant in _tenants!) {
          cards.add(TenantCard(
            tenant: tenant,
            parentCallback: refreshFromChildren,
          ));
        }
      }
      if (_tools != null && _tools!.isNotEmpty) {
        _hasOpenDcim = false;
        _hasNetbox = false;
        for (var tool in _tools!) {
          var type = Tools.netbox;
          if (tool.name.contains(Tools.opendcim.name)) {
            type = Tools.opendcim;
            _hasOpenDcim = true;
          } else {
            _hasNetbox = true;
          }
          cards.add(ToolCard(
            type: type,
            container: tool,
            parentCallback: refreshFromChildren,
          ));
        }
      }
    } else {
      cards.add(AutoProjectCard(
        namespace: Namespace.Physical,
        userEmail: widget.userEmail,
        parentCallback: refreshFromChildren,
      ));
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
