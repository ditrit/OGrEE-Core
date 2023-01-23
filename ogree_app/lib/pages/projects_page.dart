import 'package:flutter/material.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/pages/select_page.dart';

class ProjectsPage extends StatelessWidget {
  const ProjectsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: myAppBar(),
      body: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 80.0, vertical: 20),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Mes projets',
                    style: Theme.of(context).textTheme.headlineLarge),
                Padding(
                  padding: const EdgeInsets.only(right: 10.0, bottom: 10),
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => const SelectPage(),
                        ),
                      );
                    },
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: const [
                        Padding(
                          padding: EdgeInsets.symmetric(vertical: 10.0),
                          child: Icon(Icons.add_to_photos),
                        ),
                        Text(
                          "   Créer un nouveau projet",
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
            Expanded(
              child: GridView.extent(
                padding: const EdgeInsets.only(top: 5),
                maxCrossAxisExtent: 270,
                children: getProjectCards(),
              ),
            ),
          ],
        ),
      ),
    );
  }

  getProjectCards() {
    List<Card> cards = [];
    for (var i = 1; i <= 9; i++) {
      cards.add(Card(
        margin: const EdgeInsets.all(10),
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text("Projet ABC",
                      style:
                          TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  Icon(
                    Icons.star,
                    color: i == 1 ? Colors.blueAccent : Colors.grey,
                    size: 20,
                  )
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Auteur :"),
                  ),
                  Text(
                    " DOE John ",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2.0),
                    child: Text("Dernière modification :"),
                  ),
                  Text(
                    " 12/10/21 ",
                    style: TextStyle(backgroundColor: Colors.grey.shade200),
                  ),
                ],
              ),
              Align(
                alignment: Alignment.bottomRight,
                child: TextButton.icon(
                    onPressed: () {},
                    icon: const Icon(Icons.play_circle),
                    label: const Text("Lancer")),
              )
            ],
          ),
        ),
      ));
    }
    return cards;
  }
}
