import 'package:flex_color_picker/flex_color_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/theme.dart';

class CustomFormField extends StatefulWidget {
  Function(String?) save;
  String label;
  IconData icon;
  List<TextInputFormatter>? formatters;
  String? initialValue;
  bool shouldValidate;
  bool isColor;
  TextEditingController? colorTextController;
  int? maxLength;
  bool isCompact;
  String tipStr;
  String? prefix;
  String? suffix;
  bool isObscure;
  bool isReadOnly;
  Function(String)? extraValidationFn;
  TextEditingController? checkListController;
  List<String>? checkListValues;

  CustomFormField({
    super.key,
    required this.save,
    required this.label,
    required this.icon,
    this.formatters,
    this.initialValue,
    this.shouldValidate = true,
    this.isColor = false,
    this.colorTextController,
    this.maxLength,
    this.isCompact = false,
    this.tipStr = "",
    this.prefix,
    this.suffix,
    this.isObscure = false,
    this.isReadOnly = false,
    this.extraValidationFn,
    this.checkListController,
    this.checkListValues,
  });

  @override
  State<CustomFormField> createState() => _CustomFormFieldState();
}

class _CustomFormFieldState extends State<CustomFormField> {
  bool isSmallDisplay = false;
  Color? _localColor;
  final _defaultColor = Colors.grey.shade500;
  List<String> selectedCheckListItems = [];

  @override
  void initState() {
    super.initState();
    if (widget.initialValue != null && widget.isColor) {
      widget.colorTextController!.text = widget.initialValue!;
    }
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    final FocusNode focusNode = FocusNode();
    Widget? iconWidget;
    if (widget.checkListController != null) {
      iconWidget = PopupMenuButton<String>(
        tooltip: localeMsg.selectionOptions,
        offset: const Offset(0, -32),
        itemBuilder: (_) => checkListPopupItems(widget.checkListValues!),
        onCanceled: () {
          final String content = selectedCheckListItems.toString();
          widget.checkListController!.text =
              content.substring(1, content.length - 1).replaceAll(" ", "");
        },
        icon: const Icon(Icons.add),
      );
    }
    if (widget.isColor) {
      if (widget.colorTextController!.text != "") {
        _localColor =
            Color(int.parse("0xFF${widget.colorTextController!.text}"));
      }
      iconWidget = Padding(
        padding: const EdgeInsets.only(top: 9, bottom: 9, left: 14, right: 14),
        child: ColorIndicator(
          width: 22,
          height: 14,
          borderRadius: 4,
          color: _localColor ?? _defaultColor,
          onSelectFocus: false,
          onSelect: () async {
            // Store current color before we open the dialog.
            final Color colorBeforeDialog = _localColor ?? _defaultColor;
            final wasNull = _localColor == null;
            // Wait for the picker to close, if dialog was dismissed,
            // then restore the color we had before it was opened.
            if (!(await colorPickerDialog(
              localeMsg,
              widget.colorTextController!,
            ))) {
              setState(() {
                if (!wasNull) {
                  _localColor = colorBeforeDialog;
                  widget.colorTextController!.text =
                      colorBeforeDialog.toString().substring(10, 16);
                } else {
                  _localColor = null;
                  widget.colorTextController!.text = "";
                }
              });
            } else if (widget.colorTextController!.text == "") {
              // color confirmed but not sent to controller
              // means that it was colorBeforeDialog
              widget.colorTextController!.text =
                  colorBeforeDialog.toString().substring(10, 16);
            }
          },
        ),
      );
    }
    return Padding(
      padding: FormInputPadding,
      child: Tooltip(
        message: widget.tipStr,
        child: TextFormField(
          focusNode: focusNode,
          obscureText: widget.isObscure,
          readOnly: widget.isReadOnly,
          controller: widget.isColor
              ? widget.colorTextController
              : widget.checkListController,
          onChanged: widget.isColor
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
          onSaved: (newValue) => widget.save(newValue),
          validator: (text) {
            if (widget.shouldValidate) {
              if (text == null || text.isEmpty) {
                focusNode.requestFocus();
                return localeMsg.mandatoryField;
              }
              if (widget.isColor && text.length < 6) {
                return localeMsg.shouldHaveXChars(6);
              }
              if (widget.extraValidationFn != null) {
                return widget.extraValidationFn!(text);
              }
            }
            return null;
          },
          maxLength: widget.maxLength,
          inputFormatters: widget.isColor
              ? [FilteringTextInputFormatter.allow(RegExp('[0-9a-fA-F]'))]
              : widget.formatters,
          initialValue: widget.isColor ? null : widget.initialValue,
          decoration: GetFormInputDecoration(
            isSmallDisplay || widget.isCompact,
            widget.label,
            icon: widget.icon,
            iconColor: widget.isColor ? _localColor : null,
            iconWidget: iconWidget,
            prefixText: widget.prefix,
            suffixText: widget.suffix,
            isCompact: widget.isCompact,
          ),
          cursorWidth: 1.3,
          style: const TextStyle(fontSize: 14),
        ),
      ),
    );
  }

  Future<bool> colorPickerDialog(
    AppLocalizations localeMsg,
    TextEditingController colorTextController,
  ) async {
    return ColorPicker(
      color: _localColor ?? _defaultColor,
      onColorChanged: (Color color) => setState(() {
        colorTextController.text = color.toString().substring(10, 16);
        _localColor = color;
      }),
      borderRadius: 4,
      spacing: 5,
      runSpacing: 5,
      wheelDiameter: 155,
      padding: const EdgeInsets.only(top: 16, right: 16, left: 16),
      enableShadesSelection: false,
      heading: Text(
        localeMsg.selectColor,
        style: Theme.of(context).textTheme.titleSmall,
      ),
      showColorCode: true,
      colorCodeHasColor: true,
      colorCodePrefixStyle:
          Theme.of(context).textTheme.bodySmall!.copyWith(fontSize: 9),
      materialNameTextStyle: Theme.of(context).textTheme.bodySmall,
      colorNameTextStyle: Theme.of(context).textTheme.bodySmall,
      colorCodeTextStyle: Theme.of(context).textTheme.bodySmall,
      pickersEnabled: const <ColorPickerType, bool>{
        ColorPickerType.both: false,
        ColorPickerType.primary: true,
        ColorPickerType.accent: false,
        ColorPickerType.bw: false,
        ColorPickerType.custom: false,
        ColorPickerType.wheel: true,
      },
      pickerTypeLabels: <ColorPickerType, String>{
        ColorPickerType.primary: localeMsg.colorPrimary,
        ColorPickerType.wheel: localeMsg.colorWheel,
      },
    ).showPickerDialog(
      context,
      backgroundColor: Colors.white,
      actionsPadding: const EdgeInsets.only(bottom: 10, right: 22),
      transitionBuilder: (
        BuildContext context,
        Animation<double> a1,
        Animation<double> a2,
        Widget widget,
      ) {
        final double curvedValue =
            Curves.easeInOutBack.transform(a1.value) - 1.0;
        return Transform(
          transform: Matrix4.translationValues(0.0, curvedValue * 200, 0.0),
          child: Opacity(
            opacity: a1.value,
            child: widget,
          ),
        );
      },
      transitionDuration: const Duration(milliseconds: 400),
      constraints:
          const BoxConstraints(minHeight: 290, minWidth: 300, maxWidth: 320),
    );
  }

  List<PopupMenuEntry<String>> checkListPopupItems(List<String> allItems) {
    return allItems.map((String key) {
      return PopupMenuItem(
        padding: EdgeInsets.zero,
        height: 0,
        value: key,
        child: StatefulBuilder(
          builder: (context, localSetState) {
            return CheckboxListTile(
              controlAffinity: ListTileControlAffinity.leading,
              title: Text(key),
              value: selectedCheckListItems.contains(key),
              dense: true,
              onChanged: (bool? value) {
                setState(() {
                  if (value!) {
                    selectedCheckListItems.add(key);
                  } else {
                    selectedCheckListItems.remove(key);
                  }
                });
                localSetState(() {});
              },
            );
          },
        ),
      );
    }).toList();
  }
}
