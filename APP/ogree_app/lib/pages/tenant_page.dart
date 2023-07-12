import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/tenants/api_stats_view.dart';
import 'package:ogree_app/widgets/tenants/locked_view.dart';
import 'package:ogree_app/widgets/tenants/popups/domain_popup.dart';
import 'package:ogree_app/widgets/tenants/docker_view.dart';
import 'package:ogree_app/widgets/select_objects/app_controller.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/settings_view.dart';
import 'package:ogree_app/widgets/select_objects/tree_view/custom_tree_view.dart';
import 'package:ogree_app/widgets/tenants/user_view.dart';

class TenantPage extends StatefulWidget {
  final String userEmail;
  final Tenant? tenant;
  const TenantPage({super.key, required this.userEmail, this.tenant});

  @override
  State<TenantPage> createState() => _TenantPageState();
}

class _TenantPageState extends State<TenantPage> with TickerProviderStateMixin {
  late TabController _tabController;
  late final AppController appController = AppController();
  bool _isLocked = true;
  bool _reloadDomains = false;
  bool _isSmallDisplay = false;

  @override
  void initState() {
    super.initState();
    _tabController =
        TabController(length: widget.tenant != null ? 4 : 3, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Scaffold(
        backgroundColor: const Color.fromARGB(255, 238, 238, 241),
        appBar: myAppBar(context, widget.userEmail, isTenantMode: true),
        body: Padding(
          padding: const EdgeInsets.all(20.0),
          child: CustomScrollView(slivers: [
            SliverFillRemaining(
              hasScrollBody: false,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 14, left: 5),
                    child: Row(
                      children: [
                        IconButton(
                            onPressed: () =>
                                Navigator.of(context).push(MaterialPageRoute(
                                  builder: (context) => ProjectsPage(
                                      userEmail: widget.userEmail,
                                      isTenantMode: widget.tenant != null),
                                )),
                            icon: Icon(
                              Icons.arrow_back,
                              color: Colors.blue.shade900,
                            )),
                        const SizedBox(width: 5),
                        Text(
                          "Tenant $tenantName",
                          style: Theme.of(context).textTheme.headlineLarge,
                        ),
                      ],
                    ),
                  ),
                  Card(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        TabBar(
                          controller: _tabController,
                          dividerColor: Colors.white,
                          labelPadding:
                              const EdgeInsets.only(left: 20, right: 20),
                          labelColor: Colors.blue.shade900,
                          unselectedLabelColor: Colors.grey,
                          labelStyle: TextStyle(
                              fontSize: 14,
                              fontFamily: GoogleFonts.inter().fontFamily),
                          unselectedLabelStyle: TextStyle(
                              fontSize: 14,
                              fontFamily: GoogleFonts.inter().fontFamily),
                          isScrollable: true,
                          indicatorSize: TabBarIndicatorSize.label,
                          tabs: getTabs(localeMsg),
                        ),
                        const Divider(),
                        Container(
                          padding: const EdgeInsets.only(left: 20),
                          height: MediaQuery.of(context).size.height -
                              (isSmallDisplay ? 310 : 250),
                          width: double.maxFinite,
                          child: TabBarView(
                            physics: const NeverScrollableScrollPhysics(),
                            controller: _tabController,
                            children: getTabViews(localeMsg),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            )
          ]),
        ));
  }

  List<Tab> getTabs(AppLocalizations localeMsg) {
    List<Tab> tabs = [
      const Tab(
        text: "API Stats",
        icon: Icon(Icons.info),
      ),
      Tab(
        text: localeMsg.domain + "s",
        icon: const Icon(Icons.business),
      ),
      Tab(
        text: localeMsg.user + "s",
        icon: const Icon(Icons.person),
      ),
    ];
    if (widget.tenant != null) {
      tabs.insert(
          0,
          Tab(
            text: localeMsg.deployment,
            icon: const Icon(Icons.dns),
          ));
    }
    return tabs;
  }

  List<Widget> getTabViews(AppLocalizations localeMsg) {
    List<Widget> views = [
      _isLocked && widget.tenant != null
          ? LockedView(tenant: widget.tenant!, parentCallback: unlockView)
          : ApiStatsView(),
      _isLocked && widget.tenant != null
          ? LockedView(tenant: widget.tenant!, parentCallback: unlockView)
          : domainView(localeMsg),
      _isLocked && widget.tenant != null
          ? LockedView(tenant: widget.tenant!, parentCallback: unlockView)
          : UserView(),
    ];
    if (widget.tenant != null) {
      views.insert(0, DockerView(tName: widget.tenant!.name));
    }
    return views;
  }

  unlockView() {
    setState(() {
      _isLocked = false;
    });
  }

  domainView(AppLocalizations localeMsg) {
    return Stack(children: [
      AppControllerScope(
        controller: appController,
        child: FutureBuilder<void>(
          future: appController.init({},
              onlyDomain: true, reload: _reloadDomains, isTenantMode: true),
          builder: (_, __) {
            if (_reloadDomains) {
              _reloadDomains = false;
            }
            if (appController.isInitialized) {
              return Stack(children: [
                const CustomTreeView(isTenantMode: true),
                _isSmallDisplay ? Container() : const Align(
                  alignment: Alignment.topRight,
                  child: Padding(
                    padding: EdgeInsets.only(right: 16),
                    child: SizedBox(
                        width: 320,
                        height: 116,
                        child: Card(
                            // color: Color.fromARGB(255, 250, 253, 255),
                            child: SettingsView(isTenantMode: true))),
                  ),
                ),
              ]);
            }
            return const Center(child: CircularProgressIndicator());
          },
        ),
      ),
      Align(
        alignment: Alignment.bottomRight,
        child: Padding(
          padding: const EdgeInsets.only(bottom: 20, right: 20),
          child: ElevatedButton.icon(
            onPressed: () =>
                showCustomPopup(context, DomainPopup(parentCallback: () {
              setState(() {
                _reloadDomains = true;
              });
            })),
            icon: const Icon(Icons.add),
            label: Text("${localeMsg.create} ${localeMsg.domain}"),
          ),
        ),
      ),
    ]);
  }
}
