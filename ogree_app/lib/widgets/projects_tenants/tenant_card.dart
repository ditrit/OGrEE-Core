import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/widgets/projects_tenants/view_tenant_popup.dart';
import 'package:url_launcher/url_launcher.dart';

class TenantCard extends StatelessWidget {
  final Tenant tenant;
  const TenantCard({Key? key, required this.tenant});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
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
                    child: Text(tenant.name,
                        overflow: TextOverflow.clip,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16)),
                  ),
                  CircleAvatar(
                    radius: 13,
                    child: IconButton(
                        splashRadius: 18,
                        iconSize: 14,
                        padding: const EdgeInsets.all(2),
                        onPressed: () => viewTenantPopup(context, tenant),
                        icon: const Icon(
                          Icons.search,
                        )),
                  )
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text("API URL:"),
                  ),
                  Text(
                    "http://localhost:${tenant.apiUrl}",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text("Web URL:"),
                  ),
                  Text(
                    "http://localhost:${tenant.webUrl}",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: TextButton.icon(
                    onPressed: () {
                      launchUrl(Uri.parse("http://localhost:${tenant.webUrl}"));
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
