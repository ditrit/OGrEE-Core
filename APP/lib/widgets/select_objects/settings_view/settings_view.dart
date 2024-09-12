import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/select_objects/tree_view/tree_node.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

part '_actions.dart';
part '_advanced_find_field.dart';
part '_find_node_field.dart';
part '_header.dart';
part '_selected_chips.dart';

const Duration kAnimationDuration = Duration(milliseconds: 300);

const Color kDarkBlue = Color(0xff1565c0);

class SettingsView extends StatelessWidget {
  final bool isTenantMode;
  final Namespace namespace;
  const SettingsView(
      {super.key, required this.isTenantMode, required this.namespace,});

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    if (isTenantMode) {
      return ListView(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
        children: const [
          _Actions(isTenantMode: true),
          SizedBox(height: 8),
          _FindNodeField(),
        ],
      );
    } else {
      return TweenAnimationBuilder<double>(
        duration: kAnimationDuration,
        tween: Tween<double>(begin: .3, end: 1),
        builder: (_, double opacity, Widget? child) {
          return AnimatedOpacity(
            opacity: opacity,
            duration: kAnimationDuration,
            child: child,
          );
        },
        child: ListView(
          padding: const EdgeInsets.only(left: 16),
          children: [
            const SelectedChips(),
            const SettingsHeader(text: 'Actions'),
            const _Actions(isTenantMode: false),
            const SizedBox(height: 8),
            SettingsHeader(text: localeMsg.searchById),
            const _FindNodeField(),
            const SizedBox(height: 8),
            SettingsHeader(text: localeMsg.searchAdvanced),
            _AdvancedFindField(
              namespace: namespace,
            ),
            const SizedBox(height: 8),
            if (namespace != Namespace.Physical) Container() else const TreeFilter(),
          ],
        ),
      );
    }
  }
}
