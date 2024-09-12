part of 'settings_view.dart';

class SettingsHeader extends StatelessWidget {
  const SettingsHeader({super.key, required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder(
      duration: const Duration(milliseconds: 500),
      curve: Curves.fastOutSlowIn,
      tween: Tween<double>(begin: 40, end: 4),
      builder: (_, double padding, Widget? child) {
        return AnimatedPadding(
          padding: const EdgeInsets.all(8).copyWith(left: padding),
          duration: kAnimationDuration,
          curve: Curves.fastLinearToSlowEaseIn,
          child: child,
        );
      },
      child: Text(
        text,
        style: GoogleFonts.inter(
          fontSize: 15,
          color: Colors.black,
          fontWeight: FontWeight.w400,
        ),
      ),
    );
  }
}
