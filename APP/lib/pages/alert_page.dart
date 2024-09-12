import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:ogree_app/widgets/select_objects/settings_view/tree_filter.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

class AlertPage extends StatefulWidget {
  final String userEmail;
  final List<Alert> alerts;
  const AlertPage({super.key, required this.userEmail, required this.alerts});

  @override
  State<AlertPage> createState() => AlertPageState();

  static AlertPageState? of(BuildContext context) =>
      context.findAncestorStateOfType<AlertPageState>();
}

class AlertPageState extends State<AlertPage> with TickerProviderStateMixin {
  late final TreeAppController appController = TreeAppController();
  int selectedIndex = 0;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
        backgroundColor: const Color.fromARGB(255, 238, 238, 241),
        appBar: myAppBar(context, widget.userEmail),
        body: Padding(
          padding: const EdgeInsets.all(20.0),
          child: CustomScrollView(slivers: [
            SliverFillRemaining(
              hasScrollBody: false,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 10, left: 5),
                    child: Row(
                      children: [
                        IconButton(
                            onPressed: () =>
                                Navigator.of(context).push(MaterialPageRoute(
                                  builder: (context) => ProjectsPage(
                                      userEmail: widget.userEmail,
                                      isTenantMode: false,),
                                ),),
                            icon: Icon(
                              Icons.arrow_back,
                              color: Colors.blue.shade900,
                            ),),
                        const SizedBox(width: 5),
                        Text(
                          localeMsg.myAlerts,
                          style: Theme.of(context).textTheme.headlineLarge,
                        ),
                      ],
                    ),
                  ),
                  Card(
                    child: widget.alerts.isEmpty
                        ? SizedBox(
                            height: MediaQuery.of(context).size.height > 205
                                ? MediaQuery.of(context).size.height - 205
                                : MediaQuery.of(context).size.height,
                            child: Center(
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    Icons.thumb_up,
                                    size: 50,
                                    color: Colors.grey.shade600,
                                  ),
                                  Padding(
                                    padding: const EdgeInsets.only(top: 16),
                                    child: Text(
                                        "${AppLocalizations.of(context)!.noAlerts} :)",),
                                  ),
                                ],
                              ),
                            ),
                          )
                        : alertView(localeMsg),
                  ),
                ],
              ),
            ),
          ],),
        ),);
  }

  Column alertView(AppLocalizations localeMsg) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding:
              const EdgeInsets.only(left: 15, right: 15, top: 10, bottom: 10),
          height:
              MediaQuery.of(context).size.height - (isSmallDisplay ? 310 : 160),
          width: double.maxFinite,
          child: Row(
            children: <Widget>[
              Expanded(
                flex: 3,
                child: ListView(
                  children: getListTitles(),
                ),
              ),
              const VerticalDivider(
                width: 30,
                thickness: 0.5,
                color: Colors.grey,
              ),
              Expanded(
                flex: 7,
                child: Padding(
                  padding: const EdgeInsets.only(top: 4.0),
                  child: Align(
                    alignment: Alignment.topLeft,
                    child: ListView(children: [
                      getAlertText(selectedIndex),
                      const SizedBox(height: 20),
                      Directionality(
                        textDirection: TextDirection.rtl,
                        child: Align(
                          alignment: Alignment.topLeft,
                          child: ElevatedButton.icon(
                              onPressed: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (context) => SelectPage(
                                      project: Project(
                                          "auto",
                                          "",
                                          Namespace.Physical.name,
                                          localeMsg.autoGenerated,
                                          "auto",
                                          false,
                                          false,
                                          false,
                                          [],
                                          [widget.alerts[selectedIndex].id],
                                          [],
                                          isImpact: true,),
                                      userEmail: widget.userEmail,
                                    ),
                                  ),
                                );
                              },
                              icon: const Icon(Icons.arrow_back),
                              label: Text(localeMsg.goToImpact),),
                        ),
                      ),
                    ],),
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  List<ListTile> getListTitles() {
    final List<ListTile> list = [];
    int index = 0;
    for (final alert in widget.alerts) {
      list.add(getListTitle(
          alert.type.toUpperCase(), Colors.amber, alert.title, index,),);
      index++;
    }
    return list;
  }

  ListTile getListTitle(String type, Color typeColor, String title, int index) {
    return ListTile(
      isThreeLine: true,
      selectedColor: Colors.white,
      selectedTileColor: Colors.blue,
      selected: selectedIndex == index,
      minVerticalPadding: 10,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
      ),
      title: Row(
        children: [
          SizedBox(
            width: 80,
            height: 20,
            child: Badge(
              backgroundColor: typeColor,
              label: Text(
                " $type ",
                style: const TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,),
              ),
            ),
          ),
        ],
      ),
      subtitle: Padding(
        padding: const EdgeInsets.only(top: 6.0),
        child: Text(title),
      ),
      onTap: () {
        setState(() {
          selectedIndex = index;
        });
      },
    );
  }

  Text getAlertText(int index) {
    return Text.rich(
      TextSpan(
        children: [
          TextSpan(
            style: const TextStyle(
              fontSize: 14.0,
              color: Colors.black,
            ),
            children: [
              TextSpan(
                text: '${widget.alerts[index].type.capitalize()} Alert\n',
                style: Theme.of(context).textTheme.headlineLarge,
              ),
              TextSpan(
                  text:
                      "\n${widget.alerts[index].title}. ${widget.alerts[index].subtitle}.",),
            ],
          ),
        ],
      ),
    );
  }

  WidgetSpan getWidgetSpan(String text, String badge, MaterialColor badgeColor) {
    return WidgetSpan(
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 2.0),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              height: 20,
              child: Badge(
                backgroundColor: badgeColor.shade50,
                label: Text(
                  " $badge ",
                  style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: badgeColor.shade900,),
                ),
              ),
            ),
            Text(text),
          ],
        ),
      ),
    );
  }

  List<TextSpan> getSubtitle(String subtitle) {
    return [
      TextSpan(
        text: '\n$subtitle\n',
        style: Theme.of(context).textTheme.headlineMedium,
      ),
      const TextSpan(text: '\n'),
    ];
  }
}
