import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';

class ConfirmPopup extends StatefulWidget {
  final String objName;
  final bool isStart;
  final Function parentCallback;
  const ConfirmPopup(
      {super.key,
      required this.objName,
      required this.parentCallback,
      required this.isStart});

  @override
  State<ConfirmPopup> createState() => _ConfirmPopupState();
}

class _ConfirmPopupState extends State<ConfirmPopup> {
  bool _isLoading = false;
  bool _shouldStart = false;
  String _updateResult = "";
  final ScrollController _outputController = ScrollController();

  @override
  void initState() {
    super.initState();
    if (widget.isStart) {
      _shouldStart = true;
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    if (_shouldStart) {
      submitStopStartTenant(localeMsg, context, widget.objName);
      _shouldStart = false;
    }
    return Center(
      child: Container(
        width: 480,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 20),
          child: Material(
            color: Colors.white,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(localeMsg.areYouSure,
                    style: Theme.of(context).textTheme.headlineMedium),
                widget.isStart
                    ? Container()
                    : Padding(
                        padding: const EdgeInsets.symmetric(vertical: 30),
                        child: Text("This tenant will be completely stopped"),
                      ),
                widget.isStart
                    ? Container()
                    : Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton.icon(
                            style: OutlinedButton.styleFrom(
                                foregroundColor: Colors.blue.shade900),
                            onPressed: () => Navigator.pop(context),
                            label: Text(localeMsg.close),
                            icon: const Icon(
                              Icons.cancel_outlined,
                              size: 16,
                            ),
                          ),
                          const SizedBox(width: 15),
                          ElevatedButton.icon(
                              style: ElevatedButton.styleFrom(
                                  backgroundColor: Colors.red.shade900),
                              onPressed: () => submitStopStartTenant(
                                  localeMsg, context, widget.objName),
                              label: Text("Stop"),
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
                                  : const Icon(Icons.stop_circle, size: 16))
                        ],
                      ),
                _updateResult != ""
                    ? Padding(
                        padding: const EdgeInsets.only(top: 12),
                        child: Container(
                          height: 110,
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(12),
                            color: Colors.black,
                          ),
                          child: Padding(
                            padding: const EdgeInsets.all(8.0),
                            child: ListView(
                              controller: _outputController,
                              children: [
                                Text(
                                  "Output:$_updateResult",
                                  style: const TextStyle(color: Colors.white),
                                ),
                              ],
                            ),
                          ),
                        ),
                      )
                    : Container(),
                widget.isStart
                    ? Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          ElevatedButton.icon(
                              onPressed: () => Navigator.pop(context),
                              label: Text("OK"),
                              icon: const Icon(Icons.check_circle, size: 16))
                        ],
                      )
                    : Container(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  submitStopStartTenant(AppLocalizations localeMsg, BuildContext popupContext,
      String tenantName) async {
    final messenger = ScaffoldMessenger.of(popupContext);
    Result<Stream<String>, Exception> result;
    if (widget.isStart) {
      result = await startTenant(tenantName);
    } else {
      result = await stopTenant(tenantName);
    }
    switch (result) {
      case Success(value: final value):
        String finalMsg = "";
        if (_updateResult.isNotEmpty) {
          _updateResult = "$_updateResult\nOutput:";
        }
        await for (var chunk in value) {
          // Process each chunk as it is received
          print(chunk);
          var newLine = chunk.split("data:").last.trim();
          if (newLine.isNotEmpty) {
            setState(() {
              _updateResult = "$_updateResult\n$newLine";
              if (_outputController.hasClients) {
                _outputController
                    .jumpTo(_outputController.position.maxScrollExtent + 20);
              }
            });
          }
          if (!chunk.contains("data:")) {
            // not from the stream of events
            finalMsg = chunk;
          }
        }
        if (finalMsg.contains("Error")) {
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, "$finalMsg. Check output log below.",
              isError: true);
        } else {
          widget.parentCallback();
          if (context.mounted) {
            showSnackBar(
                ScaffoldMessenger.of(context),
                widget.isStart
                    ? "Tenant successfully started 🥳"
                    : "Tenant successfully stopped",
                isSuccess: true);
          }
          if (popupContext.mounted) Navigator.of(popupContext).pop();
        }
      case Failure(exception: final exception):
        setState(() {
          _isLoading = false;
        });
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }
}
