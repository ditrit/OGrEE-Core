import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/pages/unity_page.dart';

class AutoUnityProjectCard extends StatelessWidget {
  final String userEmail;
  const AutoUnityProjectCard({super.key, required this.userEmail});
  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    const color = Colors.blue;

    return SizedBox(
      width: 265,
      height: 250,
      child: Card(
        elevation: 3,
        surfaceTintColor: Colors.white,
        margin: const EdgeInsets.all(10),
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SizedBox(
                height: 30,
                child: Badge(
                  backgroundColor: color.shade50,
                  label: Text(
                    " 3D View ",
                    style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: color.shade900,),
                  ),
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text(localeMsg.author),
                  ),
                  Text(
                    localeMsg.autoGenerated,
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 2.0),
                    child: Text(localeMsg.descriptionTwoPoints),
                  ),
                  Text(
                    localeMsg.view3Dobjs,
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: TextButton.icon(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => UnityPage(userEmail: userEmail),
                        ),
                      );
                    },
                    icon: const Icon(Icons.play_circle),
                    label: Text(localeMsg.launch),),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
