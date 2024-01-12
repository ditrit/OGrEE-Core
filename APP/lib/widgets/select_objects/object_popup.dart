import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';

class CreateObjectPopup extends StatefulWidget {
  Function() parentCallback;
  String? parentId;
  CreateObjectPopup({super.key, required this.parentCallback, this.parentId});

  @override
  State<CreateObjectPopup> createState() => _CreateObjectPopupState();
}

enum AuthOption { pKey, password }

enum ObjCategories { site, building, room, rack, device, group }

class _CreateObjectPopupState extends State<CreateObjectPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _parentId;
  bool _isLoading = false;
  bool _isSmallDisplay = false;
  ObjCategories _objCategory = ObjCategories.site;
  List<Widget> attributesRows = [];
  List<String> attributes = [];
  Map<ObjCategories, List<String>> categoryAttrs = {};
  List<String> domainList = [];
  Map<String, dynamic> createObjData = {};
  Map<String, String> createObjDataAttrs = {};

  @override
  void initState() {
    // TODO: implement initState
    super.initState();
    if (widget.parentId != null) {
      _parentId = widget.parentId;
      switch (".".allMatches(_parentId!).length) {
        case 0:
          _objCategory = ObjCategories.building;
        case 1:
          _objCategory = ObjCategories.room;
        case 2:
          _objCategory = ObjCategories.rack;
        default:
          _objCategory = ObjCategories.device;
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

    print(_parentId);

    return FutureBuilder(
        future: categoryAttrs.isEmpty ? getExternalAssets() : null,
        builder: (context, _) {
          if (categoryAttrs.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          attributes = categoryAttrs[_objCategory]!;
          return Center(
            child: Container(
              width: 500,
              constraints: BoxConstraints(maxHeight: 590),
              margin: const EdgeInsets.symmetric(horizontal: 20),
              decoration: PopupDecoration,
              child: Padding(
                padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
                child: Form(
                    key: _formKey,
                    child: ScaffoldMessenger(
                        child: Builder(
                      builder: (context) => Scaffold(
                        backgroundColor: Colors.white,
                        body: ListView(
                          padding: EdgeInsets.zero,
                          children: [
                            Center(
                              child: Text(
                                "Créer un nouveau objet",
                                style:
                                    Theme.of(context).textTheme.headlineMedium,
                              ),
                            ),
                            // const Divider(height: 45),
                            const SizedBox(height: 20),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text("Type d'objet :"),
                                const SizedBox(width: 20),
                                SizedBox(
                                  height: 35,
                                  width: 147,
                                  child: DropdownButtonFormField<ObjCategories>(
                                    borderRadius: BorderRadius.circular(12.0),
                                    decoration: GetFormInputDecoration(
                                      false,
                                      null,
                                      icon: Icons.catching_pokemon,
                                    ),
                                    value: _objCategory,
                                    items: ObjCategories.values
                                        .map<DropdownMenuItem<ObjCategories>>(
                                            (ObjCategories value) {
                                      return DropdownMenuItem<ObjCategories>(
                                        value: value,
                                        child: Text(
                                          value.name,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      );
                                    }).toList(),
                                    onChanged: (ObjCategories? value) {
                                      setState(() {
                                        _objCategory = value!;
                                      });
                                    },
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 10),
                            _objCategory != ObjCategories.site
                                ? getFormField(
                                    save: (newValue) {
                                      if (newValue != null &&
                                          newValue.isNotEmpty) {
                                        _parentId = newValue;
                                        createObjData["parentId"] = _parentId;
                                      }
                                    },
                                    label: "Parent ID",
                                    icon: Icons.family_restroom,
                                    initial: widget.parentId)
                                : Container(),
                            getFormField(
                                save: (newValue) =>
                                    createObjData["name"] = newValue,
                                label: "Name",
                                icon: Icons.edit),
                            domainList.isEmpty
                                ? getFormField(
                                    save: (newValue) =>
                                        createObjData["domain"] = newValue,
                                    label: "Domain",
                                    icon: Icons.edit)
                                : domainAutoFillField(),
                            getFormField(
                                save: (newValue) =>
                                    createObjData["description"] = [newValue],
                                label: "Description",
                                icon: Icons.edit,
                                shouldValidate: false),
                            getFormField(
                                save: (newValue) {
                                  var tags = newValue!.split(",");
                                  if (!(tags.length == 1 && tags.first == "")) {
                                    createObjData["tags"] = tags;
                                  }
                                },
                                label: "Tags",
                                icon: Icons.tag_sharp,
                                shouldValidate: false),
                            Padding(
                              padding: const EdgeInsets.only(
                                  top: 4.0, left: 6, bottom: 6),
                              child: Text("Attributes:"),
                            ),
                            SizedBox(
                              height: (attributes.length ~/ 2 +
                                      attributes.length % 2) *
                                  60,
                              child: GridView.count(
                                physics: NeverScrollableScrollPhysics(),
                                childAspectRatio: 3.5,
                                shrinkWrap: true,
                                padding: EdgeInsets.only(left: 4),
                                // Create a grid with 2 columns
                                crossAxisCount: 2,
                                children:
                                    List.generate(attributes.length, (index) {
                                  return getFormField(
                                      save: (newValue) {
                                        if (attributes[index] == "template") {
                                          createObjDataAttrs[attributes[index]
                                                  .replaceFirst("*", "")] =
                                              newValue ?? "";
                                        } else if (newValue != null &&
                                            newValue.isNotEmpty) {
                                          createObjDataAttrs[attributes[index]
                                                  .replaceFirst("*", "")] =
                                              newValue;
                                        }
                                      },
                                      label: attributes[index],
                                      icon: Icons.tag_sharp,
                                      isCompact: true,
                                      shouldValidate:
                                          attributes[index].contains("*"));
                                }),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.only(left: 4),
                              child: Column(children: attributesRows),
                            ),
                            Padding(
                              padding: const EdgeInsets.only(left: 6),
                              child: Align(
                                alignment: Alignment.bottomLeft,
                                child: TextButton.icon(
                                    onPressed: () => setState(() {
                                          attributesRows.add(addCustomAttrRow(
                                              attributesRows.length));
                                        }),
                                    icon: const Icon(Icons.add),
                                    label: Text("Attribute")),
                              ),
                            ),

                            const SizedBox(height: 12),
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
                                    onPressed: () =>
                                        submitCreateObject(localeMsg),
                                    label: Text(localeMsg.create),
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
                          ],
                        ),
                      ),
                    ))),
              ),
            ),
          );
        });
  }

  getExternalAssets() async {
    await readJsonAssets();
    await getDomains();
  }

  readJsonAssets() async {
    for (var obj in ObjCategories.values) {
      print(obj);
      String schemaPrefix = obj.name;
      String data = await DefaultAssetBundle.of(context)
          .loadString("../API/models/schemas/${schemaPrefix}_schema.json");
      final Map<String, dynamic> jsonResult = json.decode(data);
      if (jsonResult["properties"]["attributes"]["properties"] != null) {
        var attrs = Map<String, dynamic>.from(
            jsonResult["properties"]["attributes"]["properties"]);
        print(attrs.keys);
        categoryAttrs[obj] = attrs.keys.toList();
        // categoryAttrs[obj]!.sort((a, b) => a.compareTo(b));
        if (jsonResult["properties"]["attributes"]["required"] != null) {
          var requiredAttrs = List<String>.from(
              jsonResult["properties"]["attributes"]["required"]);
          for (var i = 0; i < categoryAttrs[obj]!.length; i++) {
            var attr = categoryAttrs[obj]![i];
            if (requiredAttrs.contains(attr) && attr != "template") {
              categoryAttrs[obj]![i] = "*$attr";
            }
          }
        }
        categoryAttrs[obj]!.sort((a, b) => a.compareTo(b));
      }
    }
  }

  getDomains() async {
    final messenger = ScaffoldMessenger.of(context);
    var result = await fetchObjectsTree(
        namespace: Namespace.Organisational, isTenantMode: false);
    switch (result) {
      case Success(value: final listValue):
        domainList = listValue[0]
            .values
            .reduce((value, element) => List.from(value + element));
        print(domainList);
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
        if (context.mounted) Navigator.pop(context);
        return;
    }
  }

  domainAutoFillField() {
    return Padding(
      padding: const EdgeInsets.only(right: 10, left: 1, bottom: 6),
      child: RawAutocomplete<String>(
        optionsBuilder: (TextEditingValue textEditingValue) {
          return domainList.where((String option) {
            return option.contains(textEditingValue.text);
          });
        },
        fieldViewBuilder: (BuildContext context,
            TextEditingController textEditingController,
            FocusNode focusNode,
            VoidCallback onFieldSubmitted) {
          return TextFormField(
            controller: textEditingController,
            focusNode: focusNode,
            decoration:
                GetFormInputDecoration(false, "Domain", icon: Icons.edit),
            onFieldSubmitted: (String value) {
              createObjData["domain"] = value;
              onFieldSubmitted();
            },
            onSaved: (newValue) => createObjData["domain"] = newValue,
            onTap: () {
              // force call optionsBuilder for
              // when widgets.options changes
              // textEditingController.notifyListeners();
            },
          );
        },
        optionsViewBuilder: (BuildContext context,
            AutocompleteOnSelected<String> onSelected,
            Iterable<String> options) {
          return Align(
            alignment: Alignment.topLeft,
            child: Material(
              elevation: 4.0,
              borderRadius: BorderRadius.circular(12),
              child: SizedBox(
                height: options.length > 2 ? 150.0 : 50.0 * options.length,
                width: 400,
                child: ListView.builder(
                  padding: const EdgeInsets.all(8.0),
                  itemCount: options.length,
                  itemBuilder: (BuildContext context, int index) {
                    final String option = options.elementAt(index);
                    return GestureDetector(
                      onTap: () {
                        onSelected(option);
                      },
                      child: ListTile(
                        title:
                            Text(option, style: const TextStyle(fontSize: 14)),
                      ),
                    );
                  },
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  addCustomAttrRow(int rowIdx, {bool useDefaultValue = true}) {
    return StatefulBuilder(builder: (context, localSetState) {
      String? attrName;
      return Padding(
        padding: const EdgeInsets.only(top: 2.0),
        child: SizedBox(
          height: 60,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.start,
            children: [
              Flexible(
                flex: 5,
                child: getFormField(
                    save: (newValue) => attrName = newValue,
                    label: "Attribute",
                    icon: Icons.tag_sharp,
                    isCompact: true),
              ),
              Padding(
                padding: EdgeInsets.only(right: 6),
                child: Icon(
                  Icons.arrow_forward,
                  color: Colors.blue.shade600,
                ),
              ),
              Flexible(
                flex: 4,
                child: getFormField(
                    save: (newValue) =>
                        createObjDataAttrs[attrName!] = newValue!,
                    label: "Value",
                    icon: Icons.tag_sharp,
                    isCompact: true),
              ),
              IconButton(
                  padding: const EdgeInsets.only(bottom: 6),
                  constraints: const BoxConstraints(),
                  iconSize: 14,
                  onPressed: () =>
                      setState(() => attributesRows.removeAt(rowIdx)),
                  icon: Icon(
                    Icons.delete,
                    color: Colors.red.shade400,
                  )),
            ],
          ),
        ),
      );
    });
  }

  submitCreateObject(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });

      createObjData["category"] = _objCategory.name;
      createObjData["attributes"] = createObjDataAttrs;
      print(createObjData);

      final messenger = ScaffoldMessenger.of(context);
      final result = await createObject(createObjData, _objCategory.name);
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.createOK, isSuccess: true);
          if (context.mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      String? prefix,
      String? suffix,
      List<TextInputFormatter>? formatters,
      String? initial,
      bool shouldValidate = true,
      bool isCompact = false}) {
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
        initialValue: initial,
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (shouldValidate) {
            if (text == null || text.isEmpty) {
              return AppLocalizations.of(context)!.mandatoryField;
            }
          }
          return null;
        },
        inputFormatters: formatters,
        decoration: GetFormInputDecoration(_isSmallDisplay | isCompact, label,
            prefixText: prefix, suffixText: suffix, icon: icon),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
