import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/alert_page.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/projects/autoproject_card.dart';
import 'package:ogree_app/widgets/projects/autounity_card.dart';
import 'package:ogree_app/widgets/projects/project_card.dart';
import 'package:ogree_app/widgets/tenants/popups/create_tenant_popup.dart';
import 'package:ogree_app/widgets/tenants/tenant_card.dart';
import 'package:ogree_app/widgets/tools/create_netbox_popup.dart';
import 'package:ogree_app/widgets/tools/create_opendcim_popup.dart';
import 'package:ogree_app/widgets/tools/download_tool_popup.dart';
import 'package:ogree_app/widgets/tools/tool_card.dart';

class ProjectsPage extends StatefulWidget {
  final String userEmail;
  final bool isTenantMode;
  const ProjectsPage(
      {super.key, required this.userEmail, required this.isTenantMode,});

  @override
  State<ProjectsPage> createState() => _ProjectsPageState();
}

class _ProjectsPageState extends State<ProjectsPage> {
  List<Project>? _projects;
  List<Tenant>? _tenants;
  List<DockerContainer>? _tools;
  bool _isSmallDisplay = false;
  bool _hasNetbox = false;
  bool _hasNautobot = false;
  bool _hasOpenDcim = false;
  bool _gotData = false;
  bool _gotAlerts = false;
  List<Alert> _alerts = [];

  @override
  Widget build(BuildContext context) {
    _isSmallDisplay = MediaQuery.of(context).size.width < 720;
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: myAppBar(context, widget.userEmail,
          isTenantMode: widget.isTenantMode,),
      body: Padding(
        padding: EdgeInsets.symmetric(
            horizontal: _isSmallDisplay ? 40 : 80.0, vertical: 20,),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ...getAlertWidgets(localeMsg),
            // SizedBox(height: 30),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                if (widget.isTenantMode) Row(
                        children: [
                          Text(localeMsg.applications,
                              style: Theme.of(context).textTheme.headlineLarge,),
                          IconButton(
                              onPressed: () => setState(() {
                                    _gotData = false;
                                  }),
                              icon: const Icon(Icons.refresh),),
                        ],
                      ) else Text(localeMsg.myprojects,
                        style: Theme.of(context).textTheme.headlineLarge,),
                Row(
                  children: [
                    if (!widget.isTenantMode) Padding(
                            padding:
                                const EdgeInsets.only(right: 10.0, bottom: 10),
                            child: impactViewButton(),
                          ) else Container(),
                    Padding(
                      padding: const EdgeInsets.only(right: 10.0, bottom: 10),
                      child: createProjectButton(),
                    ),
                    if (widget.isTenantMode) Padding(
                            padding:
                                const EdgeInsets.only(right: 10.0, bottom: 10),
                            child: createToolsButton(),
                          ) else Container(),
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
                  } else if (!widget.isTenantMode) {
                    return Expanded(
                      child: SingleChildScrollView(
                        child: Wrap(
                          spacing: 5,
                          children: getCards(context),
                        ),
                      ),
                    );
                  } else {
                    if ((_tenants != null && _tenants!.isNotEmpty) ||
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
                  }
                },),
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
          final (tenants, tools) = value;
          _tenants = tenants;
          for (final tenant in tenants) {
            final result = await fetchTenantDockerInfo(tenant.name);
            switch (result) {
              case Success(value: final value):
                final List<DockerContainer> dockerInfo = value;
                if (dockerInfo.isEmpty) {
                  tenant.status = TenantStatus.unavailable;
                } else {
                  int runCount = 0;
                  for (final container in dockerInfo) {
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

  getAlerts() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchAlerts();
    switch (result) {
      case Success(value: final value):
        _alerts = value;
        setState(() {
          _gotAlerts = true;
        });
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        _projects = [];
    }
  }

  ElevatedButton createProjectButton() {
    final localeMsg = AppLocalizations.of(context)!;
    return ElevatedButton(
      onPressed: () {
        if (widget.isTenantMode) {
          showCustomPopup(
              context, CreateTenantPopup(parentCallback: refreshFromChildren),);
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
                top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10,),
            child: const Icon(Icons.add_to_photos),
          ),
          if (_isSmallDisplay) Container() else Text(widget.isTenantMode
                  ? "${localeMsg.create} tenant"
                  : localeMsg.newProject,),
        ],
      ),
    );
  }

  ElevatedButton createToolsButton() {
    final localeMsg = AppLocalizations.of(context)!;
    final List<PopupMenuEntry<Tools>> entries = <PopupMenuEntry<Tools>>[
      PopupMenuItem(
        value: Tools.netbox,
        child: Text("${localeMsg.create} Netbox"),
      ),
      PopupMenuItem(
        value: Tools.nautobot,
        child: Text("${localeMsg.create} Nautobot"),
      ),
      PopupMenuItem(
        value: Tools.opendcim,
        child: Text("${localeMsg.create} OpenDCIM"),
      ),
      PopupMenuItem(
        value: Tools.cli,
        child: Text(localeMsg.downloadCli),
      ),
      PopupMenuItem(
        value: Tools.unity,
        child: Text(localeMsg.downloadUnity),
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
                    localeMsg.onlyOneTool("Netbox"),);
              } else {
                showCustomPopup(
                    context,
                    CreateNboxPopup(
                        parentCallback: refreshFromChildren,
                        tool: Tools.netbox,),);
              }
            case Tools.nautobot:
              if (_hasNautobot) {
                showSnackBar(ScaffoldMessenger.of(context),
                    localeMsg.onlyOneTool("Nautobot"),);
              } else {
                showCustomPopup(
                    context,
                    CreateNboxPopup(
                        parentCallback: refreshFromChildren,
                        tool: Tools.nautobot,),);
              }
            case Tools.opendcim:
              if (_hasOpenDcim) {
                showSnackBar(ScaffoldMessenger.of(context),
                    localeMsg.onlyOneTool("OpenDCIM"),);
              } else {
                showCustomPopup(context,
                    CreateOpenDcimPopup(parentCallback: refreshFromChildren),);
              }
            case Tools.cli:
              showCustomPopup(context, const DownloadToolPopup(tool: Tools.cli),
                  isDismissible: true,);
            case Tools.unity:
              showCustomPopup(context, const DownloadToolPopup(tool: Tools.unity),
                  isDismissible: true,);
          }
        },
        itemBuilder: (_) => entries,
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: EdgeInsets.only(
                  top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10,),
              child: const Icon(Icons.timeline),
            ),
            if (_isSmallDisplay) Container() else Text(localeMsg.tools),
          ],
        ),
      ),
    );
  }

  List<Widget> getCards(context) {
    final List<Widget> cards = [];
    if (widget.isTenantMode) {
      if (_tenants != null && _tenants!.isNotEmpty) {
        for (final tenant in _tenants!) {
          cards.add(TenantCard(
            tenant: tenant,
            parentCallback: refreshFromChildren,
          ),);
        }
      }
      if (_tools != null && _tools!.isNotEmpty) {
        _hasOpenDcim = false;
        _hasNetbox = false;
        _hasNautobot = false;
        for (final tool in _tools!) {
          var type = Tools.netbox;
          if (tool.name.contains(Tools.opendcim.name)) {
            type = Tools.opendcim;
            _hasOpenDcim = true;
          } else if (tool.name.contains(Tools.nautobot.name)) {
            type = Tools.nautobot;
            _hasNautobot = true;
          } else {
            _hasNetbox = true;
          }
          cards.add(ToolCard(
            type: type,
            container: tool,
            parentCallback: refreshFromChildren,
          ),);
        }
      }
    } else {
      if (isDemo) {
        cards.add(AutoUnityProjectCard(
          userEmail: widget.userEmail,
        ),);
      }
      for (final namespace in Namespace.values) {
        if (namespace != Namespace.Test) {
          cards.add(AutoProjectCard(
            namespace: namespace,
            userEmail: widget.userEmail,
            parentCallback: refreshFromChildren,
          ),);
        }
      }
      for (final project in _projects!) {
        cards.add(ProjectCard(
          project: project,
          userEmail: widget.userEmail,
          parentCallback: refreshFromChildren,
        ),);
      }
    }
    return cards;
  }

  ElevatedButton impactViewButton() {
    final localeMsg = AppLocalizations.of(context)!;
    return ElevatedButton(
      style: ElevatedButton.styleFrom(
          // backgroundColor: Colors.blue.shade600,
          // foregroundColor: Colors.white,
          ),
      onPressed: () => Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => SelectPage(
            isImpact: true,
            userEmail: widget.userEmail,
          ),
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Padding(
            padding: EdgeInsets.only(
                top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10,),
            child: const Icon(Icons.settings_suggest),
          ),
          if (_isSmallDisplay) Container() else Text(localeMsg.impactAnalysis),
        ],
      ),
    );
  }

  List<Widget> getAlertWidgets(AppLocalizations localeMsg) {
    if (widget.isTenantMode) {
      return [];
    }
    return [
      Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(localeMsg.myAlerts,
              style: Theme.of(context).textTheme.headlineLarge,),
          Padding(
            padding: const EdgeInsets.only(right: 10.0, bottom: 10),
            child: alertViewButton(),
          ),
        ],
      ),
      const SizedBox(height: 20),
      Padding(
        padding: const EdgeInsets.only(right: 20.0),
        child: FutureBuilder(
            future: _gotAlerts ? null : getAlerts(),
            builder: (context, _) {
              if (!_gotAlerts) {
                return const Center(child: CircularProgressIndicator());
              }
              return InkWell(
                onTap: () => Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => AlertPage(
                      userEmail: widget.userEmail,
                      alerts: _alerts,
                    ),
                  ),
                ),
                child: MaterialBanner(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 5),
                  content: _isSmallDisplay
                      ? Text(localeMsg.oneAlert)
                      : (_alerts.isEmpty
                          ? Text("${localeMsg.noAlerts} :)")
                          : Text(alertsToString(localeMsg))),
                  leading: const Icon(Icons.info),
                  backgroundColor: _alerts.isEmpty
                      ? Colors.grey.shade200
                      : Colors.amber.shade100,
                  dividerColor: Colors.transparent,
                  actions: const <Widget>[
                    TextButton(
                      onPressed: null,
                      child: Text(''),
                    ),
                  ],
                ),
              );
            },),
      ),
      const SizedBox(height: 30),
    ];
  }

  String alertsToString(AppLocalizations localeMsg) {
    var alertStr = "";
    if (_alerts.length > 1) {
      for (final alert in _alerts) {
        alertStr = "$alertStr${alert.title.split(" ").first}, ";
      }
      alertStr = alertStr.substring(0, alertStr.length - 2);
      alertStr = "${localeMsg.areMarkedMaintenance} $alertStr.";
    } else {
      for (final alert in _alerts) {
        alertStr = "${alert.title.split(" ").first} ${localeMsg.isMarked}.";
      }
    }
    return alertStr;
  }

  ElevatedButton alertViewButton() {
    final localeMsg = AppLocalizations.of(context)!;
    return ElevatedButton(
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.deepOrange,
        foregroundColor: Colors.white,
      ),
      onPressed: () {
        if (widget.isTenantMode) {
          showCustomPopup(
              context, CreateTenantPopup(parentCallback: refreshFromChildren),);
        } else {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => AlertPage(
                userEmail: widget.userEmail,
                alerts: _alerts,
              ),
            ),
          );
        }
      },
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Padding(
            padding: EdgeInsets.only(
                top: 8, bottom: 8, right: _isSmallDisplay ? 0 : 10,),
            child: const Icon(Icons.analytics),
          ),
          if (_isSmallDisplay) Container() else Text(localeMsg.viewAlerts),
        ],
      ),
    );
  }
}
