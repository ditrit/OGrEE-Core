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

class LogicalObjectPopup extends StatefulWidget {
  Function() parentCallback;
  String? objId;
  Namespace namespace;
  LogicalObjectPopup(
      {super.key,
      required this.parentCallback,
      required this.namespace,
      this.objId});

  @override
  State<LogicalObjectPopup> createState() => _LogicalObjectPopupState();
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

class _LogicalObjectPopupState extends State<LogicalObjectPopup> {
  final _formKey = GlobalKey<FormState>();
  bool _isSmallDisplay = false;
  LogCategories _objCategory = LogCategories.group;
  String _objId = "";
  List<Widget> attributesRows = [];
  List<String> attributes = [];
  Map<String, List<String>> categoryAttrs = {};
  List<String> domainList = [];
  Map<String, dynamic> createObjData = {};
  Map<String, dynamic> editObjData = {};
  Map<String, String> createObjDataAttrs = {};
  bool _isEdit = false;

  String? _tagSlug;
  String? _tagDescription;
  String? _tagColor;
  Color? _localColor;
  Tag? tag;
  PlatformFile? _loadedImage;

  PlatformFile? _loadedFile;
  String? _loadFileResult;

  @override
  void initState() {
    super.initState();
    if (widget.objId != null && widget.objId!.isNotEmpty) {
      _isEdit = true;
      _objId = widget.objId!;
    }
    print("IS EDIT LOOOOOOG " + _isEdit.toString());
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
          print(categoryAttrs);
          print(_objCategory);
          return Center(
            child: Container(
              width: 500,
              constraints: BoxConstraints(
                  maxHeight: _objCategory == LogCategories.group ? 585 : 390),
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
                            // padding: EdgeInsets.zero,
                            children: [
                              Center(
                                child: Text(
                                  _isEdit
                                      ? (_objCategory.name.contains("template")
                                          ? "Visualiser template"
                                          : "Modifier l'objet")
                                      : "Créer un nouveau objet",
                                  style: Theme.of(context)
                                      .textTheme
                                      .headlineMedium,
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
                                      isExpanded: true,
                                      borderRadius: BorderRadius.circular(12.0),
                                      decoration: GetFormInputDecoration(
                                        false,
                                        null,
                                        icon: Icons.bookmark,
                                      ),
                                      value: _objCategory.name,
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
                                                _objCategory = LogCategories
                                                    .values
                                                    .firstWhere(
                                                        (e) => e.name == value);
                                              });
                                            },
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 10),
                              SizedBox(
                                  height: _objCategory == LogCategories.group
                                      ? 415
                                      : 220,
                                  child: getFormByCategory(
                                      _objCategory, localeMsg)),

                              const SizedBox(height: 12),
                              ActionBtnRow(
                                isEdit: _isEdit,
                                onlyDelete: _isEdit &&
                                    _objCategory.name.contains("template"),
                                submitCreate: () {
                                  switch (_objCategory) {
                                    case LogCategories.tag:
                                      return onActionBtnPressed(localeMsg);
                                    case LogCategories.group:
                                    case LogCategories.layer:
                                      return submitCreateObject(
                                          localeMsg, context);
                                    default:
                                      return submitCreateTemplate(
                                          localeMsg, context);
                                  }
                                },
                                submitModify: () {
                                  switch (_objCategory) {
                                    case LogCategories.tag:
                                      return onActionBtnPressed(localeMsg);
                                    case LogCategories.group:
                                    case LogCategories.layer:
                                      return submitModifyObject(
                                          localeMsg, context);
                                    default:
                                      return;
                                  }
                                },
                                submitDelete: () =>
                                    submitDeleteObject(localeMsg, context),
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

  getExternalAssets() async {
    await readJsonAssets();
    await getDomains();
    if (_isEdit) {
      await getObject();
    }
  }

  readJsonAssets() async {
    var obj = LogCategories.group.name;
    // List<String> objects = objsByNamespace[widget.namespace]!;
    // for (var obj in objects) {
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
    var errMsg = "";
    for (var keyId in ["id", "slug"]) {
      var result = await fetchObject(_objId, idKey: keyId);
      switch (result) {
        case Success(value: final value):
          if (["room", "building", "device", "rack"]
              .contains(value["category"])) {
            switch (value["category"]) {
              case "room":
                _objCategory = LogCategories.room_template;
                break;
              case "building":
                _objCategory = LogCategories.bldg_template;
                break;
              default:
                _objCategory = LogCategories.obj_template;
            }
            var encoder = new JsonEncoder.withIndent("     ");
            _loadFileResult = encoder.convert(value);
            // _loadFileResult = value.toString();
          } else {
            print("my category ${value["category"]}");
            if (value["applicability"] != null) {
              _objCategory = LogCategories.layer;
              createObjData = value;
              createObjDataAttrs =
                  Map<String, String>.from(createObjData["filters"]);
              print(createObjDataAttrs);
              for (var attr in createObjDataAttrs.entries) {
                attributesRows.add(addCustomAttrRow(attributesRows.length,
                    givenAttrName: attr.key, givenAttrValue: attr.value));
              }
            } else if (value["category"] == null) {
              print("chaaaanging");
              _objCategory = LogCategories.tag;
              tag = Tag.fromMap(value);
              _localColor = Color(int.parse("0xFF${tag!.color}"));
            } else {
              createObjData = value;
              createObjDataAttrs =
                  Map<String, String>.from(createObjData["attributes"]);
            }

            print("GOT OBJECT");
            print(createObjDataAttrs);
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
                    save: (newValue) =>
                        createObjDataAttrs[attrName!] = newValue!,
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
                    setState(() => attributesRows.removeAt(rowIdx));
                    createObjDataAttrs.remove(attrName);
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

  getFormByCategory(LogCategories category, AppLocalizations localeMsg) {
    print("get form by " + category.name);
    switch (category) {
      case LogCategories.group:
        return getGroupForm();
      case LogCategories.layer:
        return getLayerForm();
      case LogCategories.tag:
        return getTagForm();
      default:
        return getTemplatesForm(localeMsg);
    }
  }

  getGroupForm() {
    print("HELLOOO GROOOUP");
    attributes = categoryAttrs[_objCategory.name]!;
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        getFormField(
            save: (newValue) {
              if (newValue != null && newValue.isNotEmpty) {
                createObjData["parentId"] = newValue;
              }
            },
            label: "Parent ID",
            icon: Icons.family_restroom,
            initial: createObjData["parentId"]),
        getFormField(
            save: (newValue) => createObjData["name"] = newValue,
            label: "Name",
            icon: Icons.edit,
            initial: createObjData["name"]),
        (domainList.isEmpty
            ? getFormField(
                save: (newValue) => createObjData["domain"] = newValue,
                label: "Domain",
                icon: Icons.edit,
                initial: createObjData["domain"])
            : domainAutoFillField()),
        getFormField(
            save: (newValue) => createObjData["description"] = [newValue],
            label: "Description",
            icon: Icons.edit,
            shouldValidate: false,
            initial: createObjData["description"] != null &&
                    List<String>.from(createObjData["description"]).isNotEmpty
                ? createObjData["description"][0]
                : null),
        getFormField(
            save: (newValue) {
              var tags = newValue!.replaceAll(" ", "").split(",");
              if (!(tags.length == 1 && tags.first == "")) {
                createObjData["tags"] = tags;
              }
            },
            label: "Tags",
            icon: Icons.tag_sharp,
            shouldValidate: false,
            initial: createObjData["tags"]
                ?.toString()
                .substring(1, createObjData["tags"].toString().length - 1)),
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
              print(
                  createObjDataAttrs[attributes[index].replaceFirst("*", "")]);
              return getFormField(
                  save: (newValue) {
                    if (newValue != null && newValue.isNotEmpty) {
                      createObjDataAttrs[
                          attributes[index].replaceFirst("*", "")] = newValue;
                    }
                  },
                  label: attributes[index],
                  icon: Icons.tag_sharp,
                  isCompact: true,
                  shouldValidate: attributes[index].contains("*"),
                  initial: createObjDataAttrs[
                      attributes[index].replaceFirst("*", "")]);
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
                      attributesRows
                          .add(addCustomAttrRow(attributesRows.length));
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
          save: (newValue) => createObjData["slug"] = newValue,
          label: "Name",
          icon: Icons.edit,
          initial: createObjData["slug"]),
      getFormField(
          save: (newValue) => createObjData["applicability"] = newValue,
          label: "Applicability",
          icon: Icons.edit,
          initial: createObjData["applicability"]),
      Padding(
          padding: const EdgeInsets.only(top: 4.0, left: 6, bottom: 6),
          child: Text("Filters:")),
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
                    attributesRows.add(addCustomAttrRow(attributesRows.length));
                  }),
              icon: const Icon(Icons.add),
              label: Text("Filter")),
        ),
      ),
    ]);
  }

  getTagForm() {
    final localeMsg = AppLocalizations.of(context)!;
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        getFormField(
            save: (newValue) => _tagSlug = newValue,
            label: "Slug",
            icon: Icons.auto_awesome_mosaic,
            initial: _isEdit ? tag!.slug : null),
        getFormField(
            save: (newValue) => _tagDescription = newValue,
            label: "Description",
            icon: Icons.auto_awesome_mosaic,
            initial: _isEdit ? tag!.description : null),
        getFormField(
            save: (newValue) => _tagColor = newValue,
            label: localeMsg.color,
            icon: Icons.circle,
            formatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9a-fA-F]'))
            ],
            isColor: true,
            initial: _isEdit ? tag!.color : null),
        Padding(
          padding: const EdgeInsets.only(top: 8.0, bottom: 8),
          child: Wrap(
            alignment: WrapAlignment.end,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              _loadedImage != null || (_isEdit && tag!.image != "")
                  ? IconButton(
                      padding: const EdgeInsets.all(4),
                      constraints: const BoxConstraints(),
                      iconSize: 14,
                      onPressed: () {
                        setState(() {
                          _loadedImage = null;
                          tag!.image = "";
                        });
                      },
                      icon: const Icon(
                        Icons.cancel_outlined,
                      ))
                  : Container(),
              Padding(
                padding: const EdgeInsets.only(right: 20),
                child: _loadedImage == null
                    ? (_isEdit && tag!.image != ""
                        ? Image.network(
                            tenantUrl + tag!.image,
                            height: 40,
                          )
                        : Container())
                    : Image.memory(
                        _loadedImage!.bytes!,
                        height: 40,
                      ),
              ),
              ElevatedButton.icon(
                  onPressed: () async {
                    FilePickerResult? result = await FilePicker.platform
                        .pickFiles(
                            type: FileType.custom,
                            allowedExtensions: ["png", "jpg", "jpeg", "webp"],
                            withData: true);
                    if (result != null) {
                      setState(() {
                        _loadedImage = result.files.single;
                      });
                    }
                  },
                  icon: const Icon(Icons.download),
                  label: Text(
                      _isSmallDisplay ? "Image" : "${localeMsg.select} image")),
            ],
          ),
        )
      ],
    );
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

  onActionBtnPressed(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      var newTag = Tag(
          slug: _tagSlug!,
          description: _tagDescription!,
          color: _tagColor!,
          image: _loadedImage != null
              ? "data:image/png;base64,${base64Encode(_loadedImage!.bytes!)}"
              : "");
      Result result;
      if (_isEdit) {
        var newTagMap = newTag.toMap();
        if (_loadedImage == null && tag!.image != "") {
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
  }

  submitCreateTemplate(
      AppLocalizations localeMsg, BuildContext popupContext) async {
    final messenger = ScaffoldMessenger.of(context);
    final errorMessenger = ScaffoldMessenger.of(popupContext);
    if (_loadedFile == null) {
      showSnackBar(messenger, localeMsg.mustSelectJSON);
    } else {
      var result = await createTemplate(_loadedFile!.bytes!, _objCategory.name);
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

      if (_objCategory == LogCategories.layer) {
        createObjData["filters"] = createObjDataAttrs;
      } else {
        createObjData["category"] = _objCategory.name;
        createObjData["attributes"] = createObjDataAttrs;
      }
      print(createObjData);

      final messenger = ScaffoldMessenger.of(context);
      final errorMessenger = ScaffoldMessenger.of(popupContext);
      final result = await createObject(createObjData, _objCategory.name);
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

      if (_objCategory == LogCategories.layer) {
        createObjData["filters"] = createObjDataAttrs;
      } else {
        createObjData["category"] = _objCategory.name;
        createObjData["attributes"] = createObjDataAttrs;
      }

      createObjData.remove("lastUpdated");
      createObjData.remove("createdDate");
      createObjData.remove("id");
      print(createObjData);

      final messenger = ScaffoldMessenger.of(popupContext);
      final result =
          await updateObject(_objId, _objCategory.name, createObjData);

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
    var result = await deleteObject(_objId, _objCategory.name);
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
      bool isColor = false,
      bool isCompact = false}) {
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
        initialValue: initial,
        onChanged: isColor
            ? (value) {
                if (value.length == 6) {
                  setState(() {
                    _localColor = Color(int.parse("0xFF$value"));
                  });
                } else {
                  setState(() {
                    _localColor = null;
                  });
                }
              }
            : null,
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
            prefixText: prefix,
            suffixText: suffix,
            icon: icon,
            iconColor: isColor ? _localColor : null),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
