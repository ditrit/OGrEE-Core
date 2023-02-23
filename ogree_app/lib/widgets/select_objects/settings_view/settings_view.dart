import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import '../app_controller.dart';
import 'tree_filter.dart';

part '_actions.dart';
part '_find_node_field.dart';
part '_header.dart';
part '_selected_chips.dart';

const Duration kAnimationDuration = Duration(milliseconds: 300);

const Color kDarkBlue = Color(0xff1565c0);

class SettingsView extends StatelessWidget {
  const SettingsView({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
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
          const _ActionsHeader(),
          const _Actions(),
          const SizedBox(height: 10),
          SettingsHeader(text: localeMsg.searchById),
          const _FindNodeField(),
          const SizedBox(height: 10),
          TreeFilter(),
        ],
      ),
    );
  }
}
