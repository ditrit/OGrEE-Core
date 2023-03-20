import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api.dart';

import '../../models/tenant.dart';

void viewTenantPopup(BuildContext context, Tenant tenant) {
  showGeneralDialog(
    context: context,
    barrierDismissible: true,
    barrierLabel: "Barrier",
    barrierColor: Colors.black.withOpacity(0.5),
    transitionDuration: const Duration(milliseconds: 700),
    pageBuilder: (context, _, __) {
      return ViewTenantCard(tenant: tenant);
    },
    transitionBuilder: (_, anim, __, child) {
      Tween<Offset> tween;
      if (anim.status == AnimationStatus.reverse) {
        tween = Tween(begin: const Offset(-1, 0), end: Offset.zero);
      } else {
        tween = Tween(begin: const Offset(1, 0), end: Offset.zero);
      }
      return SlideTransition(
        position: tween.animate(anim),
        child: FadeTransition(
          opacity: anim,
          child: child,
        ),
      );
    },
  );
}

class ViewTenantCard extends StatefulWidget {
  Tenant tenant;
  ViewTenantCard({super.key, required this.tenant});

  @override
  State<ViewTenantCard> createState() => _ViewTenantCardState();
}

class _ViewTenantCardState extends State<ViewTenantCard> {
  Map<String, dynamic>? _tenantStats;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    var tenant = widget.tenant;
    return Center(
      child: Container(
        height: 650,
        width: 525,
        margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 5),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Material(
            color: Colors.white,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    const Icon(Icons.info),
                    Text(
                      "  Tenant ${tenant.name}",
                      style: GoogleFonts.inter(
                        fontSize: 22,
                        color: Colors.black,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const Divider(height: 45),
                FutureBuilder(
                    future: getTenantStats(),
                    builder: (context, _) {
                      if (_tenantStats == null) {
                        return const Center(child: CircularProgressIndicator());
                      } else if (_tenantStats!.isNotEmpty) {
                        List<Widget> stats = [];
                        for (var key in _tenantStats!.keys) {
                          stats.add(Padding(
                            padding: const EdgeInsets.only(left: 2, right: 10),
                            child: Row(
                              children: [
                                Text(
                                  "$key : ",
                                  style: const TextStyle(
                                      fontWeight: FontWeight.bold),
                                ),
                                Text(_tenantStats![key].toString())
                              ],
                            ),
                          ));
                        }
                        return Expanded(
                          child: SingleChildScrollView(
                              child: Column(children: stats)),
                        );
                      } else {
                        // Empty messages
                        return Text(localeMsg.noProjects);
                      }
                    }),
                const SizedBox(height: 25),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    ElevatedButton.icon(
                        onPressed: () {
                          Navigator.of(context).pop();
                        },
                        label: const Text("OK"),
                        icon: const Icon(Icons.thumb_up, size: 16))
                  ],
                )
              ],
            ),
          ),
        ),
      ),
    );
  }

  getTenantStats() async {
    _tenantStats =
        await fetchTenantStats("http://localhost:${widget.tenant.apiUrl}");
    Map<String, dynamic> versionStats =
        await fetchTenantApiVersion("http://localhost:${widget.tenant.apiUrl}");
    for (var key in versionStats.keys) {
      if (key.contains("Build")) {
        _tenantStats!["API$key"] = versionStats[key];
      } else {
        _tenantStats![key] = versionStats[key];
      }
    }
  }
}
