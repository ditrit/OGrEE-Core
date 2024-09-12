import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/netbox.dart';

class DeleteDialog extends StatefulWidget {
  final List<String> objName;
  final String objType;
  final Function parentCallback;
  const DeleteDialog({
    super.key,
    required this.objName,
    required this.parentCallback,
    required this.objType,
  });

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
        height: 200,
        width: 480,
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 8),
          child: Material(
            color: Colors.white,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  localeMsg.areYouSure,
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 30),
                  child: Text(localeMsg.allWillBeLost),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton.icon(
                      style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.red.shade900,
                      ),
                      onPressed: () => deleteAction(localeMsg),
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

  deleteAction(AppLocalizations localeMsg) async {
    setState(() => _isLoading = true);
    final messenger = ScaffoldMessenger.of(context);
    for (final obj in widget.objName) {
      Result result;
      if (widget.objType == "tenants") {
        result = await deleteTenant(obj);
      } else if (Tools.values
          .where((tool) => tool.name == widget.objType)
          .isNotEmpty) {
        result = await deleteTool(widget.objType);
      } else {
        result = await removeObject(obj, widget.objType);
      }
      switch (result) {
        case Success():
          break;
        case Failure(exception: final exception):
          showSnackBar(
            messenger,
            "Error: $exception",
            isError: true,
            duration: const Duration(seconds: 30),
          );
          setState(() => _isLoading = false);
          return;
      }
    }
    setState(() => _isLoading = false);
    widget.parentCallback();
    showSnackBar(messenger, localeMsg.deleteOK);
    if (mounted) Navigator.of(context).pop();
  }
}
