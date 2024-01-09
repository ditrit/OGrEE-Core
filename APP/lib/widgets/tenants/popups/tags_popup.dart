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

class TagsPopup extends StatefulWidget {
  Function() parentCallback;
  String? tagId;
  TagsPopup({super.key, required this.parentCallback, this.tagId});

  @override
  State<TagsPopup> createState() => _TagsPopupState();
}

class _TagsPopupState extends State<TagsPopup> with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  String? _tagSlug;
  String? _tagDescription;
  String? _tagColor;
  Color? _localColor;
  bool _isLoading = false;
  bool _isLoadingDelete = false;
  bool _isEdit = false;
  Tag? tag;
  PlatformFile? _loadedImage;
  bool _isSmallDisplay = false;

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
          return TagForm(localeMsg);
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
        _localColor = Color(int.parse("0xFF${tag!.color}"));
      case Failure():
        showSnackBar(messenger, localeMsg.noDomain, isError: true);
        if (context.mounted) Navigator.of(context).pop();
        return;
    }
  }

  TagForm(AppLocalizations localeMsg) {
    return Center(
      child: Container(
        width: 500,
        constraints: const BoxConstraints(maxHeight: 400),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
              _isSmallDisplay ? 30 : 40, 8, _isSmallDisplay ? 30 : 40, 15),
          child: Material(
            color: Colors.white,
            child: Form(
              key: _formKey,
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
                    height: 270,
                    child: Padding(
                      padding: const EdgeInsets.only(top: 16.0),
                      child: getTagForm(),
                    ),
                  ),
                  const SizedBox(height: 5),
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
                      _isEdit
                          ? TextButton.icon(
                              style: OutlinedButton.styleFrom(
                                  foregroundColor: Colors.red.shade900),
                              onPressed: () => onDeleteBtnPressed(localeMsg),
                              label:
                                  Text(_isSmallDisplay ? "" : localeMsg.delete),
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
                            )
                          : Container(),
                      _isSmallDisplay ? Container() : const SizedBox(width: 10),
                      ElevatedButton.icon(
                        onPressed: () => onActionBtnPressed(localeMsg),
                        label:
                            Text(_isEdit ? localeMsg.modify : localeMsg.create),
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
                            : const Icon(Icons.check_circle, size: 16),
                      ),
                    ],
                  )
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  onDeleteBtnPressed(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoadingDelete = true;
      });
      final messenger = ScaffoldMessenger.of(context);
      var result = await removeObject(tag!.slug, "tags");
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.deleteOK);
          if (context.mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoadingDelete = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  onActionBtnPressed(AppLocalizations localeMsg) async {
    final messenger = ScaffoldMessenger.of(context);
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });
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
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
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
                  label: Text(
                      _isSmallDisplay ? "Image" : "${localeMsg.select} image")),
            ],
          ),
        )
      ],
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
          if (isColor && text.length < 6) {
            return localeMsg.shouldHaveXChars(6);
          }
          return null;
        },
        maxLength: isColor ? 6 : null,
        inputFormatters: formatters,
        initialValue: initialValue,
        decoration: GetFormInputDecoration(_isSmallDisplay, label,
            icon: icon, iconColor: isColor ? _localColor : null),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
