import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:file_picker/file_picker.dart';
import 'package:ogree_app/models/tag.dart';
import 'package:ogree_app/widgets/actionbtn_row.dart';

class TagsPopup extends StatefulWidget {
  Function() parentCallback;
  String? tagId;
  TagsPopup({super.key, required this.parentCallback, this.tagId});

  @override
  State<TagsPopup> createState() => _TagsPopupState();
}

class _TagsPopupState extends State<TagsPopup> with TickerProviderStateMixin {
  bool _isEdit = false;
  Tag? tag;
  bool _isSmallDisplay = false;
  final GlobalKey<TagFormState> _tagFormKey = GlobalKey();

  @override
  void initState() {
    super.initState();
    _isEdit = widget.tagId != null;
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);

    return FutureBuilder(
      future: _isEdit && tag == null ? getTag(localeMsg) : null,
      builder: (context, _) {
        if (!_isEdit || (_isEdit && tag != null)) {
          return getTagForm(localeMsg);
        } else {
          return const Center(child: CircularProgressIndicator());
        }
      },
    );
  }

  getTag(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    final result = await fetchTag(widget.tagId!);
    switch (result) {
      case Success(value: final value):
        tag = value;
      case Failure():
        showSnackBar(messenger, localeMsg.noDomain, isError: true);
        if (context.mounted) Navigator.of(context).pop();
        return;
    }
  }

  getTagForm(AppLocalizations localeMsg) {
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 380),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              _isSmallDisplay ? 30 : 40, 8, _isSmallDisplay ? 30 : 40, 15),
          child: Material(
            color: Colors.white,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              mainAxisSize: MainAxisSize.min,
              children: [
                const SizedBox(height: 12),
                Center(
                  child: Text(
                    _isEdit
                        ? "${localeMsg.modify} Tag"
                        : "${localeMsg.create} Tag",
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                ),
                const SizedBox(height: 10),
                SizedBox(
                  height: 250,
                  child: Padding(
                    padding: const EdgeInsets.only(top: 16.0),
                    child: TagForm(key: _tagFormKey, tag: tag),
                  ),
                ),
                const SizedBox(height: 5),
                ActionBtnRow(
                    isEdit: _isEdit,
                    submitCreate: () => onActionBtnPressed(localeMsg),
                    submitModify: () => onActionBtnPressed(localeMsg),
                    submitDelete: () => () => onDeleteBtnPressed(localeMsg)),
              ],
            ),
          ),
        ),
      ),
    );
  }

  onDeleteBtnPressed(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    var result = await removeObject(tag!.slug, "tags");
    switch (result) {
      case Success():
        widget.parentCallback();
        showSnackBar(messenger, localeMsg.deleteOK);
        if (context.mounted) Navigator.of(context).pop();
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }

  onActionBtnPressed(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    final localeMsg = AppLocalizations.of(context)!;

    Tag? newTag = _tagFormKey.currentState!.onActionBtnPressed();
    if (newTag == null) {
      return;
    }

    Result result;
    if (_isEdit) {
      var newTagMap = newTag!.toMap();
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
}

class TagForm extends StatefulWidget {
  Tag? tag;
  TagForm({super.key, this.tag});
  @override
  State<TagForm> createState() => TagFormState();
}

class TagFormState extends State<TagForm> {
  final _formKey = GlobalKey<FormState>();
  String? _tagSlug;
  String? _tagDescription;
  String? _tagColor;
  Color? _localColor;
  bool _isEdit = false;
  Tag? tag;

  PlatformFile? _loadedImage;

  bool _isSmallDisplay = false;

  @override
  void initState() {
    super.initState();

    if (widget.tag != null) {
      tag = widget.tag;
      _localColor = Color(int.parse("0xFF${tag!.color}"));
      _isEdit = true;
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Form(
      key: _formKey,
      child: ListView(
        padding: EdgeInsets.zero,
        children: [
          getFormField(
              save: (newValue) => _tagSlug = newValue,
              label: "Slug",
              icon: Icons.auto_awesome_mosaic,
              initialValue: _isEdit ? tag!.slug : null),
          getFormField(
              save: (newValue) => _tagDescription = newValue,
              label: "Description",
              icon: Icons.auto_awesome_mosaic,
              initialValue: _isEdit ? tag!.description : null),
          getFormField(
              save: (newValue) => _tagColor = newValue,
              label: localeMsg.color,
              icon: Icons.circle,
              formatters: [
                FilteringTextInputFormatter.allow(RegExp(r'[0-9a-fA-F]'))
              ],
              isColor: true,
              initialValue: _isEdit ? tag!.color : null),
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
                    label: Text(_isSmallDisplay
                        ? "Image"
                        : "${localeMsg.select} image")),
              ],
            ),
          )
        ],
      ),
    );
  }

  getFormField(
      {required Function(String?) save,
      required String label,
      required IconData icon,
      List<TextInputFormatter>? formatters,
      bool isColor = false,
      String? initialValue,
      bool noValidation = false}) {
    final localeMsg = AppLocalizations.of(context)!;
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
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
          if (noValidation) {
            return null;
          }
          if (text == null || text.isEmpty) {
            return AppLocalizations.of(context)!.mandatoryField;
          }
          if (isColor && text.length != 6) {
            return localeMsg.shouldHaveXChars(6);
          }
          return null;
        },
        inputFormatters: formatters,
        initialValue: initialValue,
        decoration: GetFormInputDecoration(_isSmallDisplay, label,
            icon: icon, iconColor: isColor ? _localColor : null),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }

  Tag? onActionBtnPressed() {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      var newTag = Tag(
          slug: _tagSlug!,
          description: _tagDescription!,
          color: _tagColor!,
          image: _loadedImage != null
              ? "data:image/png;base64,${base64Encode(_loadedImage!.bytes!)}"
              : "");
      return newTag;
    }
    return null;
  }
}
