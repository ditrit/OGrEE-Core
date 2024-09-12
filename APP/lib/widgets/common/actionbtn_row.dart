import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';

class ActionBtnRow extends StatefulWidget {
  Function() submitCreate;
  Function() submitModify;
  Function() submitDelete;
  bool isEdit;
  bool onlyDelete;

  ActionBtnRow(
      {super.key,
      required this.isEdit,
      required this.submitCreate,
      required this.submitModify,
      required this.submitDelete,
      this.onlyDelete = false,});
  @override
  State<ActionBtnRow> createState() => _ActionBtnRowState();
}

class _ActionBtnRowState extends State<ActionBtnRow> {
  bool _isSmallDisplay = false;
  bool _isLoading = false;
  bool _isLoadingDelete = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        TextButton.icon(
          style:
              OutlinedButton.styleFrom(foregroundColor: Colors.blue.shade900),
          onPressed: () => Navigator.pop(context),
          label: Text(localeMsg.cancel),
          icon: const Icon(
            Icons.cancel_outlined,
            size: 16,
          ),
        ),
        // const SizedBox(width: 15),
        if (widget.isEdit) TextButton.icon(
                style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.red.shade900,),
                onPressed: () async {
                  setState(() {
                    _isLoadingDelete = true;
                  });
                  await widget.submitDelete();
                  setState(() {
                    _isLoadingDelete = false;
                  });
                },
                label: Text(_isSmallDisplay ? "" : localeMsg.delete),
                icon: _isLoadingDelete
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
              ) else Container(),
        if (_isSmallDisplay) Container() else const SizedBox(width: 10),
        if (widget.onlyDelete) Container() else ElevatedButton.icon(
                onPressed: () async {
                  setState(() {
                    _isLoading = true;
                  });
                  widget.isEdit
                      ? await widget.submitModify()
                      : await widget.submitCreate();
                  setState(() {
                    _isLoading = false;
                  });
                },
                label:
                    Text(widget.isEdit ? localeMsg.modify : localeMsg.create),
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
                    : const Icon(Icons.check_circle, size: 16),),
      ],
    );
  }
}
