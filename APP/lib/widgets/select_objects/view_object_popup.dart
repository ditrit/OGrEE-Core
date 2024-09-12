import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/select_objects/object_popup.dart';

class ViewObjectPopup extends StatefulWidget {
  String objId;
  Namespace namespace;
  ViewObjectPopup({super.key, required this.namespace, required this.objId});

  @override
  State<ViewObjectPopup> createState() => _ViewObjectPopupState();
}

class _ViewObjectPopupState extends State<ViewObjectPopup> {
  String _objCategory = LogCategories.group.name;
  String? _loadFileResult;

  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return FutureBuilder(
      future: _loadFileResult == null ? getObject() : null,
      builder: (context, _) {
        if (_loadFileResult == null) {
          return const Center(child: CircularProgressIndicator());
        }

        return Center(
          child: Container(
            width: 500,
            constraints: const BoxConstraints(maxHeight: 430),
            margin: const EdgeInsets.symmetric(horizontal: 20),
            decoration: PopupDecoration,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
              child: ScaffoldMessenger(
                child: Builder(
                  builder: (context) => Scaffold(
                    backgroundColor: Colors.white,
                    body: SingleChildScrollView(
                      child: Column(
                        children: [
                          Center(
                            child: Text(
                              localeMsg.viewJSON,
                              style: Theme.of(context).textTheme.headlineMedium,
                            ),
                          ),
                          const SizedBox(height: 10),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(localeMsg.objType),
                              const SizedBox(width: 20),
                              SizedBox(
                                height: 35,
                                width: 147,
                                child: DropdownButtonFormField<String>(
                                  isExpanded: true,
                                  borderRadius: BorderRadius.circular(12.0),
                                  decoration: GetFormInputDecoration(
                                    false,
                                    null,
                                    icon: Icons.bookmark,
                                  ),
                                  value: _objCategory,
                                  items: getCategoryMenuItems(),
                                  onChanged: null,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 10),
                          SizedBox(height: 270, child: getViewForm(localeMsg)),
                          const SizedBox(height: 12),
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
              ),
            ),
          ),
        );
      },
    );
  }

  List<DropdownMenuItem<String>> getCategoryMenuItems() {
    return [
      DropdownMenuItem<String>(
        value: _objCategory,
        child: Text(
          _objCategory,
          overflow: TextOverflow.ellipsis,
        ),
      ),
    ];
  }

  Future<void> getObject() async {
    // Get object info for popup
    final messenger = ScaffoldMessenger.of(context);
    var errMsg = "";
    // Try both id and slug since we dont know the obj's category
    for (final keyId in ["id", "slug"]) {
      final result = await fetchObject(
        widget.objId,
        AppLocalizations.of(context)!,
        idKey: keyId,
      );
      switch (result) {
        case Success(value: final value):
          if (widget.namespace == Namespace.Logical) {
            if (value["applicability"] != null) {
              // layers
              _objCategory = LogCategories.layer.name;
            } else if (value["category"] == null) {
              // tags
              _objCategory = LogCategories.tag.name;
            } else {
              // group or virtual
              _objCategory = value["category"];
            }
          } else {
            // physical or organisational
            _objCategory = value["category"];
          }
          const encoder = JsonEncoder.withIndent("     ");
          _loadFileResult = encoder.convert(value);
          return;
        case Failure(exception: final exception):
          errMsg = exception.toString();
      }
    }
    showSnackBar(messenger, errMsg, isError: true);
    if (mounted) Navigator.pop(context);
  }

  Center getViewForm(AppLocalizations localeMsg) {
    return Center(
      child: ListView(
        shrinkWrap: true,
        children: [
          Container(
            color: Colors.black,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: SelectableText(
                _loadFileResult!,
                style: const TextStyle(color: Colors.white),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
