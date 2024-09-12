import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';

class ContainerLogsPopup extends StatefulWidget {
  final String containerName;
  const ContainerLogsPopup({super.key, required this.containerName});

  @override
  State<ContainerLogsPopup> createState() => _ContainerLogsPopupState();
}

class _ContainerLogsPopupState extends State<ContainerLogsPopup> {
  String? logs;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Center(
      child: Container(
        height: 650,
        width: 525,
        margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 5),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
        ),
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
                      "  Container ${widget.containerName}",
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
                    if (logs == null) {
                      return const Center(child: CircularProgressIndicator());
                    } else if (logs != "") {
                      return Expanded(
                        child: SingleChildScrollView(
                          child: SelectableText(logs!),
                        ),
                      );
                    } else {
                      // Empty messages
                      return Text("${localeMsg.noDockerLogs} :(");
                    }
                  },
                ),
                const SizedBox(height: 25),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    ElevatedButton.icon(
                      onPressed: () {
                        Navigator.of(context).pop();
                      },
                      label: const Text("OK"),
                      icon: const Icon(Icons.thumb_up, size: 16),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  getTenantStats() async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchContainerLogs(widget.containerName);
    switch (result) {
      case Success(value: final value):
        logs = value;
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        logs = "";
    }
  }
}
