import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/widgets/delete_dialog_popup.dart';
import 'package:ogree_app/widgets/tenants/tenant_card.dart';
import 'package:url_launcher/url_launcher.dart';

class ToolCard extends StatelessWidget {
  final DockerContainer container;
  final Function parentCallback;
  const ToolCard(
      {Key? key, required this.container, required this.parentCallback});

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
              right: 20.0, left: 20.0, top: 15, bottom: 13),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Container(
                    height: 24,
                    child: Badge(
                      backgroundColor: Colors.green.shade600,
                      label: const Text(" NETBOX "),
                    ),
                  ),
                  CircleAvatar(
                    radius: 13,
                    child: IconButton(
                        splashRadius: 18,
                        iconSize: 14,
                        padding: const EdgeInsets.all(2),
                        onPressed: () =>
                            showCustomPopup(context, const ImportNetboxPopup()),
                        icon: const Icon(
                          Icons.upload,
                        )),
                  ),
                ],
              ),
              const SizedBox(height: 1),
              Container(
                width: 145,
                child: Text(container.name,
                    overflow: TextOverflow.clip,
                    style: const TextStyle(
                        fontWeight: FontWeight.bold, fontSize: 16)),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Ports:"),
                  ),
                  Text(
                    container.ports,
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
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
                              objName: [container.name],
                              parentCallback: parentCallback,
                              objType: "netbox",
                            ),
                            isDismissible: true),
                        icon: Icon(
                          Icons.delete,
                          color: Colors.red.shade900,
                        )),
                  ),
                  TextButton.icon(
                      onPressed: () {
                        launchUrl(Uri.parse(container.ports));
                      },
                      icon: const Icon(Icons.play_circle),
                      label: Text(localeMsg.launch)),
                ],
              )
            ],
          ),
        ),
      ),
    );
  }
}

class ImportNetboxPopup extends StatefulWidget {
  const ImportNetboxPopup({super.key});

  @override
  State<ImportNetboxPopup> createState() => _ImportNetboxPopupState();
}

class _ImportNetboxPopupState extends State<ImportNetboxPopup> {
  PlatformFile? _loadedFile;
  bool _isLoading = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    var isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 240),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              isSmallDisplay ? 30 : 40, 20, isSmallDisplay ? 30 : 40, 15),
          child: ScaffoldMessenger(
              child: Builder(
                  builder: (context) => Scaffold(
                        backgroundColor: Colors.white,
                        body: ListView(shrinkWrap: true, children: [
                          Center(
                              child: Text(
                            localeMsg.importNetbox,
                            style: Theme.of(context).textTheme.headlineMedium,
                          )),
                          // const Divider(height: 35),
                          const SizedBox(height: 50),
                          Align(
                            child: ElevatedButton.icon(
                                onPressed: () async {
                                  FilePickerResult? result =
                                      await FilePicker.platform.pickFiles(
                                          type: FileType.custom,
                                          allowedExtensions: ["sql"],
                                          withData: true);
                                  if (result != null) {
                                    setState(() {
                                      _loadedFile = result.files.single;
                                    });
                                  }
                                },
                                icon: const Icon(Icons.download),
                                label: Text(localeMsg.selectSQL)),
                          ),
                          _loadedFile != null
                              ? Padding(
                                  padding: const EdgeInsets.only(
                                      top: 8.0, bottom: 8.0),
                                  child: Align(
                                    child: Text(localeMsg
                                        .fileLoaded(_loadedFile!.name)),
                                  ),
                                )
                              : Container(),
                          SizedBox(height: _loadedFile != null ? 27 : 57),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              TextButton.icon(
                                style: OutlinedButton.styleFrom(
                                    foregroundColor: Colors.blue.shade900),
                                onPressed: () => Navigator.pop(context),
                                label: Text(localeMsg.cancel),
                                icon: const Icon(
                                  Icons.cancel_outlined,
                                  size: 16,
                                ),
                              ),
                              const SizedBox(width: 15),
                              ElevatedButton.icon(
                                  onPressed: () => submitNetboxDump(localeMsg),
                                  label: const Text("OK"),
                                  icon: _isLoading
                                      ? Container(
                                          width: 24,
                                          height: 24,
                                          padding: const EdgeInsets.all(2.0),
                                          child:
                                              const CircularProgressIndicator(
                                            color: Colors.white,
                                            strokeWidth: 3,
                                          ),
                                        )
                                      : const Icon(Icons.check_circle,
                                          size: 16))
                            ],
                          )
                        ]),
                      ))),
        ),
      ),
    );
  }

  submitNetboxDump(AppLocalizations localeMsg) async {
    if (_loadedFile != null) {
      setState(() {
        _isLoading = true;
      });
      // Load dump first
      var result = await uploadNetboxDump(_loadedFile!);
      switch (result) {
        case Success():
          break;
        case Failure(exception: final exception):
          showSnackBar(context, "${localeMsg.failedToUpload} $exception");
      }

      // Import dump
      result = await importNetboxDump();
      switch (result) {
        case Success():
          showSnackBar(context, localeMsg.importNetboxOK, isSuccess: true);
          Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(context, exception.toString(), isError: true);
      }
    } else {
      showSnackBar(context, localeMsg.mustSelectFile, isError: true);
    }
  }
}
