import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/widgets/common/delete_dialog_popup.dart';
import 'package:url_launcher/url_launcher.dart';

class ToolCard extends StatelessWidget {
  final Tools type;
  final DockerContainer container;
  final Function parentCallback;
  const ToolCard({
    super.key,
    required this.type,
    required this.container,
    required this.parentCallback,
  });

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
            right: 20.0,
            left: 20.0,
            top: 15,
            bottom: 13,
          ),
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
                      backgroundColor: Colors.green.shade600,
                      label: type == Tools.opendcim
                          ? const Text(" OpenDCIM ")
                          : type == Tools.nautobot
                              ? const Text(" NAUTOBOT ")
                              : const Text(" NETBOX "),
                    ),
                  ),
                  if (type != Tools.netbox)
                    Container()
                  else
                    CircleAvatar(
                      radius: 13,
                      child: IconButton(
                        splashRadius: 18,
                        iconSize: 14,
                        padding: const EdgeInsets.all(2),
                        onPressed: () => showCustomPopup(
                          context,
                          const ImportNetboxPopup(),
                        ),
                        icon: const Icon(
                          Icons.upload,
                        ),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 1),
              SizedBox(
                width: 145,
                child: Text(
                  container.name,
                  overflow: TextOverflow.clip,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("URL:"),
                  ),
                  Text(
                    container.ports.isEmpty
                        ? localeMsg.unavailable
                        : container.ports,
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              const SizedBox(height: 2),
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 5),
                child: Row(
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
                            objType: type.name,
                          ),
                          isDismissible: true,
                        ),
                        icon: Icon(
                          Icons.delete,
                          color: Colors.red.shade900,
                        ),
                      ),
                    ),
                    SizedBox(
                      height: 26,
                      width: 26,
                      child: IconButton.filled(
                        style: IconButton.styleFrom(
                          backgroundColor: Colors.blue.shade700,
                        ),
                        // splashColor: Colors.blue,
                        padding: EdgeInsets.zero,
                        onPressed: () {
                          launchUrl(Uri.parse(container.ports));
                        },
                        iconSize: 16,
                        icon: const Icon(
                          Icons.open_in_new_rounded,
                          // color: Colors.blue.shade800,
                        ),
                        // label: Text(localeMsg.launch)
                      ),
                    ),
                  ],
                ),
              ),
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
    final isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 240),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
            isSmallDisplay ? 30 : 40,
            20,
            isSmallDisplay ? 30 : 40,
            15,
          ),
          child: ScaffoldMessenger(
            child: Builder(
              builder: (context) => Scaffold(
                backgroundColor: Colors.white,
                body: ListView(
                  shrinkWrap: true,
                  children: [
                    Center(
                      child: Text(
                        localeMsg.importNetbox,
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                    ),
                    // const Divider(height: 35),
                    const SizedBox(height: 50),
                    Align(
                      child: ElevatedButton.icon(
                        onPressed: () async {
                          final FilePickerResult? result =
                              await FilePicker.platform.pickFiles(
                            type: FileType.custom,
                            allowedExtensions: ["sql"],
                            withData: true,
                          );
                          if (result != null) {
                            setState(() {
                              _loadedFile = result.files.single;
                            });
                          }
                        },
                        icon: const Icon(Icons.download),
                        label: Text(localeMsg.selectSQL),
                      ),
                    ),
                    if (_loadedFile != null)
                      Padding(
                        padding: const EdgeInsets.only(
                          top: 8.0,
                          bottom: 8.0,
                        ),
                        child: Align(
                          child: Text(
                            localeMsg.fileLoaded(_loadedFile!.name),
                          ),
                        ),
                      )
                    else
                      Container(),
                    SizedBox(height: _loadedFile != null ? 27 : 57),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton.icon(
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.blue.shade900,
                          ),
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
                                  child: const CircularProgressIndicator(
                                    color: Colors.white,
                                    strokeWidth: 3,
                                  ),
                                )
                              : const Icon(
                                  Icons.check_circle,
                                  size: 16,
                                ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  submitNetboxDump(AppLocalizations localeMsg) async {
    if (_loadedFile != null) {
      final messenger = ScaffoldMessenger.of(context);
      setState(() {
        _isLoading = true;
      });
      // Load dump first
      var result = await uploadNetboxDump(_loadedFile!);
      switch (result) {
        case Success():
          break;
        case Failure(exception: final exception):
          showSnackBar(messenger, "${localeMsg.failedToUpload} $exception");
      }

      // Import dump
      result = await importNetboxDump();
      switch (result) {
        case Success():
          showSnackBar(messenger, localeMsg.importNetboxOK, isSuccess: true);
          if (mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    } else {
      showSnackBar(
        ScaffoldMessenger.of(context),
        localeMsg.mustSelectFile,
        isError: true,
      );
    }
  }
}
