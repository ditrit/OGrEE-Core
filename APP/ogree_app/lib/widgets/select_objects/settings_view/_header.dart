part of 'settings_view.dart';

class SettingsHeader extends StatelessWidget {
  const SettingsHeader({Key? key, required this.text}) : super(key: key);

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
          fontSize: 17,
          color: Colors.black,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }
}

class _ActionsHeader extends StatelessWidget {
  const _ActionsHeader({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Row(
      children: const [
        SettingsHeader(text: 'Actions'),
        Spacer(),
        Tooltip(
          message: 'Something?',
          verticalOffset: 13,
          decoration: BoxDecoration(
            color: Colors.blueAccent,
            borderRadius: BorderRadius.all(Radius.circular(12)),
          ),
          textStyle: TextStyle(
            fontSize: 13,
            color: Colors.white,
            letterSpacing: 1.025,
            height: 1.5,
          ),
          padding: EdgeInsets.all(16),
          child: Icon(Icons.info_outline_rounded, color: Colors.blueAccent),
        ),
        SizedBox(width: 10),
      ],
    );
  }
}
