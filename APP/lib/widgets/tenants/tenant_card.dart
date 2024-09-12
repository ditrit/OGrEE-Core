import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/tenant_page.dart';
import 'package:ogree_app/widgets/common/delete_dialog_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/confirm_popup.dart';
import 'package:ogree_app/widgets/tenants/popups/update_tenant_popup.dart';
import 'package:url_launcher/url_launcher.dart';

class TenantCard extends StatelessWidget {
  final Tenant tenant;
  final Function parentCallback;
  const TenantCard(
      {super.key, required this.tenant, required this.parentCallback,});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return SizedBox(
      width: 265,
      height: 260,
      child: Card(
        elevation: 3,
        surfaceTintColor: Colors.white,
        margin: const EdgeInsets.all(10),
        child: Padding(
          padding: const EdgeInsets.only(
              right: 20.0, left: 20.0, top: 15, bottom: 13,),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  SizedBox(
                    height: 24,
                    child: Badge(
                      backgroundColor: Colors.blue.shade600,
                      label: const Text(" TENANT "),
                    ),
                  ),
                  Row(
                    children: [
                      CircleAvatar(
                        radius: 13,
                        child: IconButton(
                            splashRadius: 18,
                            iconSize: 14,
                            padding: const EdgeInsets.all(2),
                            onPressed: () => showCustomPopup(
                                context,
                                UpdateTenantPopup(
                                  parentCallback: parentCallback,
                                  tenant: tenant,
                                ),),
                            icon: const Icon(
                              Icons.edit,
                            ),),
                      ),
                      const SizedBox(width: 8),
                      CircleAvatar(
                        radius: 13,
                        child: IconButton(
                            splashRadius: 18,
                            iconSize: 14,
                            padding: const EdgeInsets.all(2),
                            onPressed: () => Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (context) => TenantPage(
                                        userEmail: "admin", tenant: tenant,),
                                  ),
                                ),
                            icon: const Icon(
                              Icons.search,
                            ),),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 1),
              Row(
                children: [
                  Icon(Icons.circle,
                      color: getTenantStatusColor(tenant.status), size: 10,),
                  const SizedBox(width: 6),
                  SizedBox(
                    width: 160,
                    child: Text(tenant.name,
                        overflow: TextOverflow.clip,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16,),),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("API URL:"),
                  ),
                  Text(
                    "${tenant.apiUrl}:${tenant.apiPort}",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Web URL:"),
                  ),
                  if (tenant.webUrl.isEmpty && tenant.webPort.isEmpty) Text(localeMsg.notCreated,
                          style:
                              TextStyle(backgroundColor: Colors.grey.shade200),) else InkWell(
                          child: Text(
                            "${tenant.webUrl}:${tenant.webPort}",
                            style: TextStyle(
                                backgroundColor: Colors.grey.shade200,
                                color: Colors.blue,
                                decoration: TextDecoration.underline,
                                decorationColor: Colors.blue,),
                          ),
                          onTap: () => launchUrl(
                              Uri.parse("${tenant.webUrl}:${tenant.webPort}"),),
                        ),
                ],
              ),
              const SizedBox(height: 2),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  CircleAvatar(
                    backgroundColor: Colors.red.shade100,
                    radius: 13,
                    child: IconButton(
                        splashRadius: 18,
                        iconSize: 14,
                        padding: const EdgeInsets.all(2),
                        onPressed: () => showCustomPopup(
                            context,
                            DeleteDialog(
                              objName: [tenant.name],
                              parentCallback: parentCallback,
                              objType: "tenants",
                            ),
                            isDismissible: true,),
                        icon: Icon(
                          Icons.delete,
                          color: Colors.red.shade900,
                        ),),
                  ),
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    child: Row(
                      children: [
                        IconButton(
                          constraints: const BoxConstraints(),
                          padding: EdgeInsets.zero,
                          onPressed: () => showCustomPopup(
                              context,
                              ConfirmPopup(
                                parentCallback: parentCallback,
                                objName: tenant.name,
                                isStart: false,
                              ),),
                          icon: Icon(
                            Icons.stop_circle_sharp,
                            color: Colors.orange.shade800,
                            size: 28,
                          ),
                          // label: Text(localeMsg.launch)
                        ),
                        const SizedBox(width: 4),
                        IconButton(
                          constraints: const BoxConstraints(),
                          padding: EdgeInsets.zero,
                          onPressed: () => showCustomPopup(
                              context,
                              ConfirmPopup(
                                parentCallback: parentCallback,
                                objName: tenant.name,
                                isStart: true,
                              ),),
                          icon: Icon(
                            Icons.play_circle,
                            color: Colors.blue.shade700,
                            size: 28,
                          ),
                          // label: Text(localeMsg.launch)
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Color getTenantStatusColor(TenantStatus? status) {
    if (status == null || status == TenantStatus.unavailable) {
      return Colors.grey;
    } else if (status == TenantStatus.running) {
      return Colors.green;
    } else if (status == TenantStatus.partialRun) {
      return Colors.orange;
    } else {
      return Colors.red;
    }
  }
}
