import 'dart:convert';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/models/tag.dart';
import 'package:ogree_app/widgets/actionbtn_row.dart';
import 'package:ogree_app/widgets/tenants/popups/tags_popup.dart';

class ObjectPopup extends StatefulWidget {
  Function() parentCallback;
  String? objId;
  Namespace namespace;
  String? parentId;
  ObjectPopup(
      {super.key,
      required this.parentCallback,
      required this.namespace,
      this.objId,
      this.parentId});

  @override
  State<ObjectPopup> createState() => _ObjectPopupState();
}

enum PhyCategories { site, building, room, rack, device, group }

enum OrgCategories { domain }

enum LogCategories {
  group,
  layer,
  obj_template,
  room_template,
  bldg_template,
  tag
}

Map<Namespace, List<String>> objsByNamespace = {
  Namespace.Physical: PhyCategories.values.map((e) => e.name).toList(),
  Namespace.Organisational: OrgCategories.values.map((e) => e.name).toList(),
  Namespace.Logical: LogCategories.values.map((e) => e.name).toList(),
};

class _ObjectPopupState extends State<ObjectPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isSmallDisplay = false;
  String _objCategory = LogCategories.group.name;
  String _objId = "";
  List<Widget> customAttributesRows = [];
  Map<String, List<String>> categoryAttrs = {};
  Map<String, Map<String, String>> examplesAttrs = {};
  List<String> domainList = [];
  Map<String, dynamic> objData = {};
  Map<String, String> objDataAttrs = {};
  bool _isEdit = false;

  // Tags
  Tag? tag;
  final GlobalKey<TagFormState> _tagFormKey = GlobalKey();

  // Templates
  PlatformFile? _loadedFile;
  String? _loadFileResult;

  @override
  void initState() {
    super.initState();
    if (widget.parentId != null) {
      // side add button to node (parent), suggest a child category
      objData["parentId"] = widget.parentId;
      if (widget.namespace == Namespace.Organisational) {
        _objCategory = OrgCategories.domain.name;
      } else if (widget.namespace == Namespace.Logical) {
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
      // floating general add button
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

          return Center(
            child: Container(
              width: 500,
              constraints:
                  BoxConstraints(maxHeight: getPopupHeightByCategory()),
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
                        body: SingleChildScrollView(
                          child: Column(
                            children: [
                              Center(
                                child: Text(
                                  _isEdit
                                      ? (_objCategory.contains("template")
                                          ? "Visualiser template"
                                          : "Modifier l'objet")
                                      : "Créer un nouveau objet",
                                  style: Theme.of(context)
                                      .textTheme
                                      .headlineMedium,
                                ),
                              ),
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
                                      isExpanded: true,
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
                              SizedBox(
                                  height: getFormHeightByCategory(),
                                  child: getFormByCategory(
                                      _objCategory, localeMsg)),
                              const SizedBox(height: 12),
                              ActionBtnRow(
                                isEdit: _isEdit,
                                onlyDelete: _isEdit &&
                                    _objCategory.contains("template"),
                                submitCreate: () {
                                  if (_objCategory == LogCategories.tag.name) {
                                    return submitActionTag(localeMsg);
                                  } else if (_objCategory
                                      .contains("template")) {
                                    return submitCreateTemplate(
                                        localeMsg, context);
                                  } else {
                                    return submitCreateObject(
                                        localeMsg, context);
                                  }
                                },
                                submitModify: () {
                                  if (_objCategory == LogCategories.tag.name) {
                                    return submitActionTag(localeMsg);
                                  } else if (_objCategory
                                      .contains("template")) {
                                    return;
                                  } else {
                                    return submitModifyObject(
                                        localeMsg, context);
                                  }
                                },
                                submitDelete: () =>
                                    submitDeleteAny(localeMsg, context),
                              )
                            ],
                          ),
                        ),
                      ),
                    ))),
              ),
            ),
          );
        });
  }

  double getPopupHeightByCategory() {
    if (widget.namespace == Namespace.Physical ||
        _objCategory == LogCategories.group.name) {
      return 585;
    } else if (widget.namespace == Namespace.Organisational) {
      return 470;
    } else {
      // Logical, except group
      return 390;
    }
  }

  double getFormHeightByCategory() {
    if (widget.namespace == Namespace.Physical ||
        _objCategory == LogCategories.group.name) {
      return 415;
    } else if (widget.namespace == Namespace.Organisational) {
      return 300;
    } else {
      // Logical, except group
      return 220;
    }
  }

  getExternalAssets() async {
    await readJsonAssets();
    await getDomains();
    if (_isEdit) {
      await getObject();
    }
  }

  readJsonAssets() async {
    List<String> objects = [LogCategories.group.name];
    if (widget.namespace == Namespace.Physical) {
      objects = objsByNamespace[widget.namespace]!;
    } else if (widget.namespace == Namespace.Organisational) {
      objects = objsByNamespace[widget.namespace]!;
    }

    for (var obj in objects) {
      // Read JSON schema
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
        // Get examples
        var examples = List<Map<String, dynamic>>.from(jsonResult["examples"]);
        examplesAttrs[obj] =
            Map<String, String>.from(examples[0]["attributes"]);
      }
    }
  }

  getDomains() async {
    // Get domains option for dropdown menu of physical
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
    // Get object info for edit popup
    final messenger = ScaffoldMessenger.of(context);
    var errMsg = "";
    // Try both id and slug since we dont know the obj's category
    for (var keyId in ["id", "slug"]) {
      var result = await fetchObject(_objId, idKey: keyId);
      switch (result) {
        case Success(value: final value):
          if (widget.namespace == Namespace.Logical) {
            if (["room", "building", "device", "rack", "generic"]
                .contains(value["category"])) {
              // templates
              switch (value["category"]) {
                case "room":
                  _objCategory = LogCategories.room_template.name;
                  break;
                case "building":
                  _objCategory = LogCategories.bldg_template.name;
                  break;
                default:
                  _objCategory = LogCategories.obj_template.name;
              }
              var encoder = const JsonEncoder.withIndent("     ");
              _loadFileResult = encoder.convert(value);
            } else {
              if (value["applicability"] != null) {
                // layers
                _objCategory = LogCategories.layer.name;
                objData = value;
                objDataAttrs = Map<String, String>.from(objData["filters"]);
                for (var attr in objDataAttrs.entries) {
                  // add filters
                  customAttributesRows.add(addCustomAttrRow(
                      customAttributesRows.length,
                      givenAttrName: attr.key,
                      givenAttrValue: attr.value));
                }
              } else if (value["category"] == null) {
                // tags
                _objCategory = LogCategories.tag.name;
                tag = Tag.fromMap(value);
              } else {
                // group
                objData = value;
                objDataAttrs = Map<String, String>.from(objData["attributes"]);
                _objCategory = value["category"];
              }
            }
          } else {
            // physical or organisational
            objData = value;
            objDataAttrs = Map<String, String>.from(objData["attributes"]);
            _objCategory = value["category"];
          }

          return;
        case Failure(exception: final exception):
          errMsg = exception.toString();
      }
    }
    showSnackBar(messenger, errMsg, isError: true);
    if (context.mounted) Navigator.pop(context);
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
          if (objData["domain"] != null) {
            textEditingController.text = objData["domain"];
          }
          return TextFormField(
            controller: textEditingController,
            focusNode: focusNode,
            decoration:
                GetFormInputDecoration(false, "Domain", icon: Icons.edit),
            onFieldSubmitted: (String value) {
              objData["domain"] = value;
              onFieldSubmitted();
            },
            onSaved: (newValue) => objData["domain"] = newValue,
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

  addCustomAttrRow(int rowIdx,
      {bool useDefaultValue = true,
      String? givenAttrName,
      String? givenAttrValue}) {
    return StatefulBuilder(builder: (context, localSetState) {
      String? attrName;
      if (givenAttrName != null) {
        attrName = givenAttrName;
      }
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
                    isCompact: true,
                    initial: givenAttrName),
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
                    save: (newValue) => objDataAttrs[attrName!] = newValue!,
                    label: "Value",
                    icon: Icons.tag_sharp,
                    isCompact: true,
                    initial: givenAttrValue),
              ),
              IconButton(
                  padding: const EdgeInsets.only(bottom: 6),
                  constraints: const BoxConstraints(),
                  iconSize: 14,
                  onPressed: () {
                    setState(() => customAttributesRows.removeAt(rowIdx));
                    objDataAttrs.remove(attrName);
                  },
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

  getFormByCategory(String category, AppLocalizations localeMsg) {
    if (widget.namespace == Namespace.Physical ||
        widget.namespace == Namespace.Organisational ||
        category == LogCategories.group.name) {
      return getObjectForm();
    } else if (category == LogCategories.layer.name) {
      return getLayerForm();
    } else if (category == LogCategories.tag.name) {
      return TagForm(key: _tagFormKey, tag: tag);
    } else {
      //templates
      return getTemplatesForm(localeMsg);
    }
  }

  getObjectForm() {
    List<String> attributes = categoryAttrs[_objCategory]!;
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        _objCategory != PhyCategories.site.name
            ? getFormField(
                save: (newValue) {
                  if (newValue != null && newValue.isNotEmpty) {
                    objData["parentId"] = newValue;
                  }
                },
                label: "Parent ID",
                icon: Icons.family_restroom,
                initial: objData["parentId"])
            : Container(),
        getFormField(
            save: (newValue) => objData["name"] = newValue,
            label: "Name",
            icon: Icons.edit,
            initial: objData["name"]),
        _objCategory != OrgCategories.domain.name
            ? (domainList.isEmpty
                ? getFormField(
                    save: (newValue) => objData["domain"] = newValue,
                    label: "Domain",
                    icon: Icons.edit,
                    initial: objData["domain"])
                : domainAutoFillField())
            : Container(),
        getFormField(
            save: (newValue) => objData["description"] = [newValue],
            label: "Description",
            icon: Icons.edit,
            shouldValidate: false,
            initial: objData["description"] != null &&
                    List<String>.from(objData["description"]).isNotEmpty
                ? objData["description"][0]
                : null),
        _objCategory != OrgCategories.domain.name
            ? getFormField(
                save: (newValue) {
                  var tags = newValue!.replaceAll(" ", "").split(",");
                  if (!(tags.length == 1 && tags.first == "")) {
                    objData["tags"] = tags;
                  }
                },
                label: "Tags",
                icon: Icons.tag_sharp,
                shouldValidate: false,
                initial: objData["tags"]
                    ?.toString()
                    .substring(1, objData["tags"].toString().length - 1))
            : Container(),
        Padding(
          padding: const EdgeInsets.only(top: 4.0, left: 6, bottom: 6),
          child: Text("Attributes:"),
        ),
        SizedBox(
          height: (attributes.length ~/ 2 + attributes.length % 2) * 60,
          child: GridView.count(
            physics: NeverScrollableScrollPhysics(),
            childAspectRatio: 3.5,
            shrinkWrap: true,
            padding: EdgeInsets.only(left: 4),
            // Create a grid with 2 columns
            crossAxisCount: 2,
            children: List.generate(attributes.length, (index) {
              return getFormField(
                  tipStr: examplesAttrs[_objCategory]
                          ?[attributes[index].replaceFirst("*", "")] ??
                      "",
                  save: (newValue) {
                    if (newValue != null && newValue.isNotEmpty) {
                      objDataAttrs[attributes[index].replaceFirst("*", "")] =
                          newValue;
                    }
                  },
                  label: attributes[index],
                  icon: Icons.tag_sharp,
                  isCompact: true,
                  shouldValidate: attributes[index].contains("*"),
                  initial:
                      objDataAttrs[attributes[index].replaceFirst("*", "")]);
            }),
          ),
        ),
        Padding(
          padding: const EdgeInsets.only(left: 4),
          child: Column(children: customAttributesRows),
        ),
        Padding(
          padding: const EdgeInsets.only(left: 6),
          child: Align(
            alignment: Alignment.bottomLeft,
            child: TextButton.icon(
                onPressed: () => setState(() {
                      customAttributesRows
                          .add(addCustomAttrRow(customAttributesRows.length));
                    }),
                icon: const Icon(Icons.add),
                label: Text("Attribute")),
          ),
        ),
      ],
    );
  }

  getLayerForm() {
    return ListView(padding: EdgeInsets.zero, children: [
      getFormField(
          save: (newValue) => objData["slug"] = newValue,
          label: "Name",
          icon: Icons.edit,
          initial: objData["slug"]),
      getFormField(
          save: (newValue) => objData["applicability"] = newValue,
          label: "Applicability",
          icon: Icons.edit,
          initial: objData["applicability"]),
      Padding(
          padding: const EdgeInsets.only(top: 4.0, left: 6, bottom: 6),
          child: Text("Filters:")),
      Padding(
        padding: const EdgeInsets.only(left: 4),
        child: Column(children: customAttributesRows),
      ),
      Padding(
        padding: const EdgeInsets.only(left: 6),
        child: Align(
          alignment: Alignment.bottomLeft,
          child: TextButton.icon(
              onPressed: () => setState(() {
                    customAttributesRows
                        .add(addCustomAttrRow(customAttributesRows.length));
                  }),
              icon: const Icon(Icons.add),
              label: Text("Filter")),
        ),
      ),
    ]);
  }

  getTemplatesForm(AppLocalizations localeMsg) {
    return Center(
      child: ListView(shrinkWrap: true, children: [
        _loadFileResult == null
            ? Align(
                child: ElevatedButton.icon(
                    onPressed: () async {
                      FilePickerResult? result = await FilePicker.platform
                          .pickFiles(
                              type: FileType.custom,
                              allowedExtensions: ["json"],
                              withData: true);
                      if (result != null) {
                        setState(() {
                          _loadedFile = result.files.single;
                        });
                      }
                    },
                    icon: const Icon(Icons.download),
                    label: Text(localeMsg.selectJSON)),
              )
            : Container(),
        _loadedFile != null
            ? Padding(
                padding: const EdgeInsets.only(top: 8.0, bottom: 8.0),
                child: Align(
                  child: Text(localeMsg.fileLoaded(_loadedFile!.name)),
                ),
              )
            : Container(),
        _loadFileResult != null
            ? Container(
                color: Colors.black,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Text(
                    _loadFileResult!,
                    style: const TextStyle(color: Colors.white),
                  ),
                ),
              )
            : Container(),
      ]),
    );
  }

  submitActionTag(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    Tag? newTag = _tagFormKey.currentState!.onActionBtnPressed();
    if (newTag == null) {
      return;
    }
    Result result;
    if (_isEdit) {
      var newTagMap = newTag.toMap();
      if (newTag.image == "" && tag!.image != "") {
        newTagMap.remove("image"); // patch and keep old one
      }
      result = await updateTag(tag!.slug, newTagMap);
    } else {
      result = await createTag(newTag);
    }
    switch (result) {
      case Success():
        widget.parentCallback();
        showSnackBar(messenger,
            "${_isEdit ? localeMsg.modifyOK : localeMsg.createOK} 🥳",
            isSuccess: true);
        if (context.mounted) Navigator.of(context).pop();
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }

  submitCreateTemplate(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    final messenger = ScaffoldMessenger.of(context);
    final errorMessenger = ScaffoldMessenger.of(popupContext);
    if (_loadedFile == null) {
      showSnackBar(messenger, localeMsg.mustSelectJSON);
    } else {
      var result = await createTemplate(_loadedFile!.bytes!, _objCategory);
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

  submitCreateObject(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();

      if (_objCategory == LogCategories.layer.name) {
        objData["filters"] = objDataAttrs;
      } else {
        objData["category"] = _objCategory;
        objData["attributes"] = objDataAttrs;
      }
      print(objData);

      final messenger = ScaffoldMessenger.of(context);
      final errorMessenger = ScaffoldMessenger.of(popupContext);
      final result = await createObject(objData, _objCategory);
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

      if (_objCategory == LogCategories.layer.name) {
        objData["filters"] = objDataAttrs;
      } else {
        objData["category"] = _objCategory;
        objData["attributes"] = objDataAttrs;
      }

      objData.remove("lastUpdated");
      objData.remove("createdDate");
      objData.remove("id");
      print(objData);

      final messenger = ScaffoldMessenger.of(popupContext);
      final result = await updateObject(_objId, _objCategory, objData);

      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.modifyOK, isSuccess: true);
          if (context.mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  submitDeleteAny(AppLocalizations localeMsg, BuildContext popupContext) async {
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
      bool isCompact = false,
      String tipStr = ""}) {
    return Padding(
      padding: FormInputPadding,
      child: Tooltip(
        message: tipStr != "" ? "Example: $tipStr" : "",
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
          decoration: GetFormInputDecoration(
              _isSmallDisplay || isCompact, label,
              prefixText: prefix, suffixText: suffix, icon: icon),
          cursorWidth: 1.3,
          style: const TextStyle(fontSize: 14),
        ),
      ),
    );
  }
}
