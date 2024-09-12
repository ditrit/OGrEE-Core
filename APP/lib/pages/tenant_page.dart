import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';
import 'package:ogree_app/widgets/tenants/api_stats_view.dart';
import 'package:ogree_app/widgets/tenants/docker_view.dart';
import 'package:ogree_app/widgets/tenants/domain_view.dart';
import 'package:ogree_app/widgets/tenants/locked_view.dart';
import 'package:ogree_app/widgets/tenants/tags_view.dart';
import 'package:ogree_app/widgets/tenants/user_view.dart';

class TenantPage extends StatefulWidget {
  final String userEmail;
  final Tenant? tenant;
  const TenantPage({super.key, required this.userEmail, this.tenant});

  @override
  State<TenantPage> createState() => TenantPageState();

  static TenantPageState? of(BuildContext context) =>
      context.findAncestorStateOfType<TenantPageState>();
}

class TenantPageState extends State<TenantPage> with TickerProviderStateMixin {
  late TabController _tabController;
  late final TreeAppController appController = TreeAppController();
  bool _isLocked = true;
  String _domainSearch = "";

  @override
  void initState() {
    super.initState();
    _tabController =
        TabController(length: widget.tenant != null ? 5 : 4, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
        backgroundColor: const Color.fromARGB(255, 238, 238, 241),
        appBar: myAppBar(context, widget.userEmail,
            isTenantMode: widget.tenant != null,),
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
                                      isTenantMode: widget.tenant != null,),
                                ),),
                            icon: Icon(
                              Icons.arrow_back,
                              color: Colors.blue.shade900,
                            ),),
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
                          tabAlignment: TabAlignment.start,
                          controller: _tabController,
                          dividerColor: Colors.white,
                          labelPadding:
                              const EdgeInsets.only(left: 20, right: 20),
                          labelColor: Colors.blue.shade900,
                          unselectedLabelColor: Colors.grey,
                          labelStyle: TextStyle(
                              fontSize: 14,
                              fontFamily: GoogleFonts.inter().fontFamily,),
                          unselectedLabelStyle: TextStyle(
                              fontSize: 14,
                              fontFamily: GoogleFonts.inter().fontFamily,),
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
            ),
          ],),
        ),);
  }

  List<Tab> getTabs(AppLocalizations localeMsg) {
    final List<Tab> tabs = [
      Tab(
        text: localeMsg.apiStats,
        icon: const Icon(Icons.info),
      ),
      Tab(
        text: "${localeMsg.domain}s",
        icon: const Icon(Icons.business),
      ),
      Tab(
        text: "${localeMsg.user}s",
        icon: const Icon(Icons.person),
      ),
      const Tab(
        text: "Tags",
        icon: Icon(Icons.turned_in),
      ),
    ];
    if (widget.tenant != null) {
      tabs.insert(
          0,
          Tab(
            text: localeMsg.deployment,
            icon: const Icon(Icons.dns),
          ),);
    }
    return tabs;
  }

  List<Widget> getTabViews(AppLocalizations localeMsg) {
    final List<Widget> views = [
      if (_isLocked && widget.tenant != null) LockedView(tenant: widget.tenant!, parentCallback: unlockView) else const ApiStatsView(),
      if (_isLocked && widget.tenant != null) LockedView(tenant: widget.tenant!, parentCallback: unlockView) else const DomainView(),
      if (_isLocked && widget.tenant != null) LockedView(tenant: widget.tenant!, parentCallback: unlockView) else _domainSearch.isEmpty
              ? UserView()
              : // user view should filter with domain
              UserView(
                  searchField: UserSearchFields.Domain,
                  searchText: _domainSearch,
                  parentCallback: () => setState(() {
                    // child calls parent to clean it once applied
                    _domainSearch = "";
                  }),
                ),
      if (_isLocked && widget.tenant != null) LockedView(tenant: widget.tenant!, parentCallback: unlockView) else const TagsView(),
    ];
    if (_domainSearch.isNotEmpty) {
      // switch to domain page
      _tabController.animateTo(widget.tenant != null ? 3 : 2);
    }
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

  changeToUserView(String domain) {
    // add domain search so rebuild knows to switch to userview tab
    setState(() {
      _domainSearch = domain;
    });
  }
}
