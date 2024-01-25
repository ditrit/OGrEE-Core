import 'package:flex_color_picker/flex_color_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

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

  CustomFormField(
      {super.key,
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
      this.isReadOnly = false});

  @override
  State<CustomFormField> createState() => _CustomFormFieldState();
}

class _CustomFormFieldState extends State<CustomFormField> {
  bool isSmallDisplay = false;
  Color? _localColor;
  final _defaultColor = Colors.grey.shade500;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    Widget? iconWidget;
    if (widget.isColor) {
      print(widget.colorTextController!.text);
      if (widget.initialValue != null) {
        widget.colorTextController!.text = widget.initialValue!;
      }
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
            // Wait for the picker to close, if dialog was dismissed,
            // then restore the color we had before it was opened.
            if (!(await colorPickerDialog(
                localeMsg, widget.colorTextController!))) {
              setState(() {
                _localColor = colorBeforeDialog;
                widget.colorTextController!.text =
                    colorBeforeDialog.toString().substring(10, 16);
              });
            }
          },
        ),
      );
    }
    return Padding(
      padding: FormInputPadding,
      child: Tooltip(
        message: widget.tipStr != ""
            ? "${AppLocalizations.of(context)!.example} ${widget.tipStr}"
            : "",
        child: TextFormField(
          obscureText: widget.isObscure,
          readOnly: widget.isReadOnly,
          controller: widget.isColor ? widget.colorTextController : null,
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
                return localeMsg.mandatoryField;
              }
              if (widget.isColor && text.length < 6) {
                return localeMsg.shouldHaveXChars(6);
              }
            }
            return null;
          },
          maxLength: widget.maxLength,
          inputFormatters: widget.formatters,
          initialValue: widget.isColor ? null : widget.initialValue,
          decoration: GetFormInputDecoration(
              isSmallDisplay || widget.isCompact, widget.label,
              icon: widget.icon,
              iconColor: widget.isColor ? _localColor : null,
              iconWidget: iconWidget,
              prefixText: widget.prefix,
              suffixText: widget.suffix),
          cursorWidth: 1.3,
          style: const TextStyle(fontSize: 14),
        ),
      ),
    );
  }

  Future<bool> colorPickerDialog(AppLocalizations localeMsg,
      TextEditingController colorTextController) async {
    return ColorPicker(
      color: _localColor ?? _defaultColor,
      onColorChanged: (Color color) => setState(() {
        print(color.toString());
        colorTextController.text = color.toString().substring(10, 16);
        _localColor = color;
      }),
      width: 40,
      height: 40,
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
      showMaterialName: false,
      showColorName: false,
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
      transitionBuilder: (BuildContext context, Animation<double> a1,
          Animation<double> a2, Widget widget) {
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
}
