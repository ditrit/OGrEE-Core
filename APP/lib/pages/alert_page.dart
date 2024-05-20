import 'package:flutter/material.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/widgets/select_objects/treeapp_controller.dart';

// THIS IS A DEMO

class AlertPage extends StatefulWidget {
  final String userEmail;
  final Tenant? tenant;
  const AlertPage({super.key, required this.userEmail, this.tenant});

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
        appBar: myAppBar(context, widget.userEmail,
            isTenantMode: widget.tenant != null),
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
                                      isTenantMode: widget.tenant != null),
                                )),
                            icon: Icon(
                              Icons.arrow_back,
                              color: Colors.blue.shade900,
                            )),
                        const SizedBox(width: 5),
                        Text(
                          "Alerts",
                          style: Theme.of(context).textTheme.headlineLarge,
                        ),
                      ],
                    ),
                  ),
                  Card(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          padding: const EdgeInsets.only(
                              left: 15, right: 15, top: 10, bottom: 10),
                          height: MediaQuery.of(context).size.height -
                              (isSmallDisplay ? 310 : 160),
                          width: double.maxFinite,
                          child: Row(
                            children: <Widget>[
                              Expanded(
                                flex: 3,
                                child: ListView(
                                  children: [
                                    getListTitle(
                                        "MINOR",
                                        Colors.amber,
                                        "The temperature of device BASIC.A.R1.A02.chassis01 is higher than usual",
                                        0),
                                    getListTitle(
                                        "HINT",
                                        Colors.grey,
                                        "All nodes of a kubernetes cluster are on the same rack",
                                        1),
                                  ],
                                ),
                              ),
                              VerticalDivider(
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
                                    child: SingleChildScrollView(
                                      child: selectedIndex == 0
                                          ? getAlertText()
                                          : getHintText(),
                                    ),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            )
          ]),
        ));
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
                style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: Colors.black),
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

  getAlertText() {
    return Text.rich(
      TextSpan(
        children: [
          TextSpan(
            style: new TextStyle(
              fontSize: 14.0,
              color: Colors.black,
            ),
            children: [
              TextSpan(
                text: 'Minor Alert\n',
                style: Theme.of(context).textTheme.headlineLarge,
              ),
              TextSpan(text: '\nThe temperature of device '),
              TextSpan(
                  text: 'BASIC.A.R1.A02.chassis01',
                  style: new TextStyle(fontWeight: FontWeight.bold)),
              TextSpan(
                  text:
                      ' is higher than usual. This could impact the performance of your applications running in a Kubernetes cluster with nodes in this device: "my-frontend-app" and "my-backend-app".\n'),
              ...getSubtitle("Details:"),
              TextSpan(
                  text:
                      'The last measurement of temperature for the device in question reads '),
              TextSpan(
                  text: '64°C',
                  style: new TextStyle(fontWeight: FontWeight.bold)),
              TextSpan(
                  text:
                      '. The temperature recommendation for a chassis of this type is to not surpass '),
              TextSpan(
                  text: '55°C',
                  style: new TextStyle(fontWeight: FontWeight.bold)),
              TextSpan(text: '.\n'),
              ...getSubtitle("Impacted by this alert:"),
              getWidgetSpan(" BASIC.A.R1.A02.chassis01", "Physical - Device",
                  Colors.teal),
              TextSpan(text: '\n'),
              ...getSubtitle("May also be impacted:"),
              getWidgetSpan(" BASIC.A.R1.A02.chassis01.blade01",
                  "Physical - Device", Colors.teal),
              getWidgetSpan(" BASIC.A.R1.A02.chassis01.blade02",
                  "Physical - Device", Colors.teal),
              getWidgetSpan(" BASIC.A.R1.A02.chassis01.blade03",
                  "Physical - Device", Colors.teal),
              getWidgetSpan(" kubernetes-cluster.my-frontend-app",
                  "Logical - Application", Colors.deepPurple),
              getWidgetSpan(" kubernetes-cluster.my-backend-app",
                  "Logical - Application", Colors.deepPurple),
            ],
          ),
        ],
      ),
    );
  }

  getHintText() {
    return Text.rich(
      TextSpan(
        children: [
          TextSpan(
            style: new TextStyle(
              fontSize: 14.0,
              color: Colors.black,
            ),
            children: [
              TextSpan(
                text: 'Hint\n',
                style: Theme.of(context).textTheme.headlineLarge,
              ),
              TextSpan(
                  text:
                      '\nAll nodes of a kubernetes cluster are servers from the same rack.\n'),
              ...getSubtitle("Details:"),
              TextSpan(
                  text:
                      'The Kubernetes cluster "kubernetes-cluster" has the following devices as its nodes: "chassis01.blade01" , "chassis01.blade02" and "chassis01.blade03". All of these devices are in the same rack "BASIC.A.R1.A02".\n'),
              ...getSubtitle("Suggestion:"),
              TextSpan(
                  text:
                      'To limit impacts to the cluster and its applications in case of issue with this rack, consider adding a server from a different rack as a node to this cluster.\n'),
              ...getSubtitle("Impacted by this hint:"),
              getWidgetSpan(
                  " BASIC.A.R1.A02", "Physical - Device", Colors.teal),
              getWidgetSpan(" kubernetes-cluster", "Logical - Application",
                  Colors.deepPurple),
              TextSpan(text: '\n'),
              ...getSubtitle("May also be impacted:"),
              getWidgetSpan(" kubernetes-cluster.my-frontend-app",
                  "Logical - Application", Colors.deepPurple),
              getWidgetSpan(" kubernetes-cluster.my-backend-app",
                  "Logical - Application", Colors.deepPurple),
            ],
          ),
        ],
      ),
    );
  }

  getWidgetSpan(String text, String badge, MaterialColor badgeColor) {
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
                      color: badgeColor.shade900),
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
      TextSpan(text: '\n'),
    ];
  }
}
