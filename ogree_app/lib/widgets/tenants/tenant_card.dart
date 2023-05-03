import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/tenant_page.dart';
import 'package:url_launcher/url_launcher.dart';

class TenantCard extends StatelessWidget {
  final Tenant tenant;
  final Function parentCallback;
  const TenantCard(
      {Key? key, required this.tenant, required this.parentCallback});

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
                    width: 145,
                    child: Text(tenant.name,
                        overflow: TextOverflow.clip,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16)),
                  ),
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
                            isDismissible: true),
                        icon: Icon(
                          Icons.delete,
                          color: Colors.red.shade900,
                        )),
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
                                    userEmail: "admin", tenant: tenant),
                              ),
                            ),
                        icon: const Icon(
                          Icons.search,
                        )),
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
                    "http://${tenant.apiUrl}:${tenant.apiPort}",
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
                  Text(
                    "http://${tenant.webUrl}:${tenant.webPort}",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: TextButton.icon(
                    onPressed: () {
                      launchUrl(Uri.parse(
                          "http://${tenant.webUrl}:${tenant.webPort}"));
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

class DeleteDialog extends StatefulWidget {
  final List<String> objName;
  final String objType;
  final Function parentCallback;
  const DeleteDialog(
      {super.key,
      required this.objName,
      required this.parentCallback,
      required this.objType});

  @override
  State<DeleteDialog> createState() => _DeleteDialogState();
}

class _DeleteDialogState extends State<DeleteDialog> {
  bool _isLoading = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        height: 230,
        width: 480,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(40)),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 10),
          child: Material(
            color: Colors.white,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(localeMsg.areYouSure,
                    style: Theme.of(context).textTheme.headlineLarge),
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 40),
                  child: Text(localeMsg.allWillBeLost),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton.icon(
                      style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.red.shade900),
                      onPressed: () async {
                        setState(() => _isLoading = true);
                        for (var obj in widget.objName) {
                          String response;
                          if (widget.objType == "tenants") {
                            response = await deleteTenant(obj);
                          } else {
                            response = await removeObject(obj, widget.objType);
                          }
                          if (response != "") {
                            showSnackBar(context, "Error: " + response);
                            return;
                          }
                        }
                        setState(() => _isLoading = false);
                        widget.parentCallback();
                        Navigator.of(context).pop();
                        showSnackBar(context, localeMsg.deleteOK);
                      },
                      label: Text(localeMsg.delete),
                      icon: _isLoading
                          ? Container(
                              width: 24,
                              height: 24,
                              padding: const EdgeInsets.all(2.0),
                              child: CircularProgressIndicator(
                                color: Colors.red.shade900,
                                strokeWidth: 3,
                              ),
                            )
                          : const Icon(
                              Icons.delete,
                              size: 16,
                            ),
                    ),
                    const SizedBox(width: 15),
                    ElevatedButton(
                      onPressed: () => Navigator.of(context).pop(),
                      child: Text(localeMsg.cancel),
                    )
                  ],
                )
              ],
            ),
          ),
        ),
      ),
    );
  }
}
