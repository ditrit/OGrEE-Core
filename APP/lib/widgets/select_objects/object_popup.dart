import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/actionbtn_row.dart';

class CreateObjectPopup extends StatefulWidget {
  Function() parentCallback;
  String? parentId;
  String? objId;
  Namespace namespace;
  CreateObjectPopup(
      {super.key,
      required this.parentCallback,
      required this.namespace,
      this.parentId,
      this.objId});

  @override
  State<CreateObjectPopup> createState() => _CreateObjectPopupState();
}

enum PhyCategories { site, building, room, rack, device, group }

enum OrgCategories { domain }

enum LogCategories { group, layer, obj_template, room_template, tag }

Map<Namespace, List<String>> objsByNamespace = {
  Namespace.Physical: PhyCategories.values.map((e) => e.name).toList(),
  Namespace.Organisational: OrgCategories.values.map((e) => e.name).toList(),
  Namespace.Logical: LogCategories.values.map((e) => e.name).toList(),
};

class _CreateObjectPopupState extends State<CreateObjectPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isSmallDisplay = false;
  String _objCategory = PhyCategories.site.name;
  String _objId = "";
  List<Widget> attributesRows = [];
  List<String> attributes = [];
  Map<String, List<String>> categoryAttrs = {};
  List<String> domainList = [];
  Map<String, dynamic> createObjData = {};
  Map<String, dynamic> editObjData = {};
  Map<String, String> createObjDataAttrs = {};
  bool _isEdit = false;

  @override
  void initState() {
    super.initState();
    if (widget.parentId != null) {
      createObjData["parentId"] = widget.parentId;
      if (widget.namespace == Namespace.Organisational) {
        _objCategory = OrgCategories.domain.name;
      } else {
        switch (".".allMatches(widget.parentId!).length) {
          case 0:
            _objCategory = PhyCategories.building.name;
          case 1:
            _objCategory = PhyCategories.room.name;
          case 2:
            _objCategory = PhyCategories.rack.name;
          default:
            _objCategory = PhyCategories.device.name;
        }
      }
    } else {
      switch (widget.namespace) {
        case Namespace.Logical:
          _objCategory = LogCategories.group.name;
          break;
        case Namespace.Organisational:
          _objCategory = OrgCategories.domain.name;
          break;
        default:
          _objCategory = PhyCategories.site.name;
      }
    }
    if (widget.objId != null && widget.objId!.isNotEmpty) {
      _isEdit = true;
      _objId = widget.objId!;
    }
    print("IS EDIT " + _isEdit.toString());
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

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
              constraints: BoxConstraints(
                  maxHeight:
                      _objCategory != OrgCategories.domain.name ? 590 : 470),
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
                                _isEdit
                                    ? "Modifier l'objet"
                                    : "Créer un nouveau objet",
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
                                  child: DropdownButtonFormField<String>(
                                    borderRadius: BorderRadius.circular(12.0),
                                    decoration: GetFormInputDecoration(
                                      false,
                                      null,
                                      icon: Icons.bookmark,
                                    ),
                                    value: _objCategory,
                                    items: objsByNamespace[widget.namespace]!
                                        .map<DropdownMenuItem<String>>(
                                            (String value) {
                                      return DropdownMenuItem<String>(
                                        value: value,
                                        child: Text(
                                          value,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      );
                                    }).toList(),
                                    onChanged: _isEdit
                                        ? null
                                        : (String? value) {
                                            setState(() {
                                              _objCategory = value!;
                                            });
                                          },
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 10),
                            _objCategory != PhyCategories.site.name
                                ? getFormField(
                                    save: (newValue) {
                                      if (newValue != null &&
                                          newValue.isNotEmpty) {
                                        createObjData["parentId"] = newValue;
                                      }
                                    },
                                    label: "Parent ID",
                                    icon: Icons.family_restroom,
                                    initial: createObjData["parentId"])
                                : Container(),
                            getFormField(
                                save: (newValue) =>
                                    createObjData["name"] = newValue,
                                label: "Name",
                                icon: Icons.edit,
                                initial: createObjData["name"]),
                            _objCategory != OrgCategories.domain.name
                                ? (domainList.isEmpty
                                    ? getFormField(
                                        save: (newValue) =>
                                            createObjData["domain"] = newValue,
                                        label: "Domain",
                                        icon: Icons.edit,
                                        initial: createObjData["domain"])
                                    : domainAutoFillField())
                                : Container(),
                            getFormField(
                                save: (newValue) =>
                                    createObjData["description"] = [newValue],
                                label: "Description",
                                icon: Icons.edit,
                                shouldValidate: false,
                                initial: createObjData["description"] != null &&
                                        List<String>.from(
                                                createObjData["description"])
                                            .isNotEmpty
                                    ? createObjData["description"][0]
                                    : null),
                            _objCategory != OrgCategories.domain.name
                                ? getFormField(
                                    save: (newValue) {
                                      var tags = newValue!
                                          .replaceAll(" ", "")
                                          .split(",");
                                      if (!(tags.length == 1 &&
                                          tags.first == "")) {
                                        createObjData["tags"] = tags;
                                      }
                                    },
                                    label: "Tags",
                                    icon: Icons.tag_sharp,
                                    shouldValidate: false,
                                    initial: createObjData["tags"]
                                        ?.toString()
                                        .substring(
                                            1,
                                            createObjData["tags"]
                                                    .toString()
                                                    .length -
                                                1))
                                : Container(),
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
                                  print(createObjDataAttrs[
                                      attributes[index].replaceFirst("*", "")]);
                                  return getFormField(
                                      save: (newValue) {
                                        if (newValue != null &&
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
                                          attributes[index].contains("*"),
                                      initial: createObjDataAttrs[
                                          attributes[index]
                                              .replaceFirst("*", "")]);
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
                            ActionBtnRow(
                              isEdit: _isEdit,
                              submitCreate: () =>
                                  submitCreateObject(localeMsg, context),
                              submitModify: () =>
                                  submitModifyObject(localeMsg, context),
                              submitDelete: () =>
                                  submitDeleteObject(localeMsg, context),
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
    if (_isEdit) {
      await getObject();
    }
  }

  readJsonAssets() async {
    List<String> objects = objsByNamespace[widget.namespace]!;
    for (var obj in objects) {
      print(obj);
      // read JSON schema
      String data = await DefaultAssetBundle.of(context)
          .loadString("../API/models/schemas/${obj}_schema.json");
      final Map<String, dynamic> jsonResult = json.decode(data);
      if (jsonResult["properties"]["attributes"]["properties"] != null) {
        // Get all properties
        var attrs = Map<String, dynamic>.from(
            jsonResult["properties"]["attributes"]["properties"]);
        print(attrs.keys);
        categoryAttrs[obj] = attrs.keys.toList();
        if (jsonResult["properties"]["attributes"]["required"] != null) {
          // Get required ones
          var requiredAttrs = List<String>.from(
              jsonResult["properties"]["attributes"]["required"]);
          for (var i = 0; i < categoryAttrs[obj]!.length; i++) {
            var attr = categoryAttrs[obj]![i];
            if (requiredAttrs.contains(attr)) {
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

  getObject() async {
    final messenger = ScaffoldMessenger.of(context);
    var result = await fetchObject(_objId);
    switch (result) {
      case Success(value: final value):
        createObjData = value;
        createObjDataAttrs =
            Map<String, String>.from(createObjData["attributes"]);
        print(createObjDataAttrs);
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
          if (createObjData["domain"] != null) {
            textEditingController.text = createObjData["domain"];
          }
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

  submitCreateObject(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();

      createObjData["category"] = _objCategory;
      createObjData["attributes"] = createObjDataAttrs;
      print(createObjData);

      final messenger = ScaffoldMessenger.of(context);
      final errorMessenger = ScaffoldMessenger.of(popupContext);
      final result = await createObject(createObjData, _objCategory);
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.createOK, isSuccess: true);
          if (context.mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          showSnackBar(errorMessenger, exception.toString(), isError: true);
      }
    }
  }

  submitModifyObject(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();

      createObjData["category"] = _objCategory;
      createObjData["attributes"] = createObjDataAttrs;
      createObjData.remove("lastUpdated");
      createObjData.remove("createdDate");
      createObjData.remove("id");
      print(createObjData);

      final messenger = ScaffoldMessenger.of(popupContext);
      final result = await updateObject(_objId, _objCategory, createObjData);

      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.modifyOK, isSuccess: true);
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  submitDeleteObject(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    final messenger = ScaffoldMessenger.of(context);
    final errorMessenger = ScaffoldMessenger.of(popupContext);
    var result = await deleteObject(_objId, _objCategory);
    switch (result) {
      case Success():
        widget.parentCallback();
        showSnackBar(messenger, localeMsg.deleteOK);
        if (context.mounted) {
          Navigator.of(context).pop();
        }
      case Failure(exception: final exception):
        showSnackBar(errorMessenger, exception.toString(), isError: true);
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
