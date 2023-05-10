// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/models/tenant.dart';

// Define a stateful widget that displays API usage statistics for a given tenant
class ApiStatsView extends StatefulWidget {
  // The tenant to display stats for
  Tenant tenant;

  ApiStatsView({
    Key? key,
    required this.tenant,
  }) : super(key: key);

  @override
  State<ApiStatsView> createState() => _ApiStatsViewState();
}

// Define the state for the ApiStatsView widget
class _ApiStatsViewState extends State<ApiStatsView> {
  // Holds the API usage statistics data for the current tenant
  Map<String, dynamic>? _tenantStats;

  @override
  Widget build(BuildContext context) {
    // Localized messages
    final localeMsg = AppLocalizations.of(context)!;
    print("HELLOOOOOO");
    return FutureBuilder(
        // Async method that fetches the tenant's API usage statistics
        future: getTenantStats(),
        builder: (context, _) {
          // If the statistics data is still being fetched, show a loading indicator
          if (_tenantStats == null) {
            return const Center(child: CircularProgressIndicator());
          }
          // If the statistics data is available and not empty, display it
          else if (_tenantStats!.isNotEmpty) {
            // Create a list of widgets to hold the statistics
            List<Widget> stats = [];

            // Iterate over each key-value pair in the statistics data
            for (var key in _tenantStats!.keys) {
              // Create a Row widget containing a label and a value for each pair
              stats.add(Padding(
                padding: const EdgeInsets.only(left: 2, right: 10),
                child: Row(
                  children: [
                    // Display the key as a bold label
                    Text(
                      key,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    // Display the value as text
                    Text(_tenantStats![key].toString())
                  ],
                ),
              ));
            }

            // Display the statistics in a column within a scrollable view
            return Expanded(
              child: SingleChildScrollView(child: Column(children: stats)),
            );
          }
          // If the statistics data is empty, display a message
          else {
            return Text(localeMsg.noProjects);
          }
        });
  }

  // Async method that fetches the tenant's API usage statistics
  getTenantStats() async {
    print("YOOOOO");
    // Fetch the statistics data from the tenant's API backend
    _tenantStats = await fetchTenantStats(
        "http://${widget.tenant.apiUrl}:${widget.tenant.apiPort}");

    // Fetch additional version information about the tenant's API
    Map<String, dynamic> versionStats = await fetchTenantApiVersion(
        "http://${widget.tenant.apiUrl}:${widget.tenant.apiPort}");

    // Append the version information to the statistics data
    for (var key in versionStats.keys) {
      if (key.contains("Build")) {
        _tenantStats!["API$key"] = versionStats[key];
      } else {
        _tenantStats![key] = versionStats[key];
      }
    }
  }
}
