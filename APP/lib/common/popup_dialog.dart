import 'package:flutter/material.dart';

void showCustomPopup(BuildContext context, Widget child,
    {isDismissible = false,}) {
  showGeneralDialog(
    context: context,
    barrierLabel: "Barrier",
    barrierDismissible: isDismissible,
    barrierColor: Colors.black.withOpacity(0.1),
    transitionDuration: const Duration(milliseconds: 700),
    pageBuilder: (context, _, __) {
      return child;
    },
    transitionBuilder: (_, anim, __, child) {
      Tween<Offset> tween;
      if (anim.status == AnimationStatus.reverse) {
        tween = Tween(begin: const Offset(-1, 0), end: Offset.zero);
      } else {
        tween = Tween(begin: const Offset(1, 0), end: Offset.zero);
      }
      return SlideTransition(
        position: tween.animate(anim),
        child: FadeTransition(
          opacity: anim,
          child: child,
        ),
      );
    },
  );
}
