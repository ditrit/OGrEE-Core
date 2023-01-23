import 'package:flutter/material.dart';
import 'package:ogree_app/pages/select_page.dart';

// Test page
class SelectAttributes extends StatefulWidget {
  List<String> selectedObjects;

  SelectAttributes({super.key, required this.selectedObjects});
  @override
  State<SelectAttributes> createState() => _SelectAttributesState();
}

List<String> attrActions = <String>['Afficher', 'Somme', 'Moyenne'];

class _SelectAttributesState extends State<SelectAttributes> {
  late List<String?> dropdownValue;
  late List<String> _selectedAttrs;

  // Sample attributes
  Map<String, bool> values = {
    "height": false,
    "heightU": false,
    "heightUnit": false,
    "model": false,
    "orientation": false,
    "posXY": false,
    "posXYUnit": false,
    "posZ": false,
    "posZUnit": false,
    "serial": false,
    "size": false,
    "sizeUnit": false,
    "template": false,
    "type": false,
    "vendor": false
  };

  Widget button(i) {
    return StatefulBuilder(builder: (context, setState) {
      return Padding(
        padding: const EdgeInsets.all(8.0),
        child: DropdownButtonHideUnderline(
          child: DropdownButton<String>(
            value: dropdownValue[i],
            isExpanded: true,
            alignment: AlignmentDirectional.center,
            onChanged: (String? value) {
              setState(() {
                dropdownValue[i] = value!;
              });
            },
            items: attrActions.map<DropdownMenuItem<String>>((String value) {
              return DropdownMenuItem<String>(
                value: value,
                child: Text(value),
              );
            }).toList(),
          ),
        ),
      );
    });
  }

  Widget label(String label) {
    return Center(
      key: Key(label),
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Text(
          label,
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),
    );
  }

  List<Widget> labelRow = [];
  List<Widget> actionRow = [];

  @override
  void initState() {
    labelRow = [label("Attributes")];
    actionRow = [label("Actions")];
    dropdownValue = List.filled(10, attrActions.first, growable: true);
    _selectedAttrs = SelectPage.of(context)!.selectedAttrs;
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    print('BUILD ATTRS');
    void runFilter(String enteredKeyword) {}

    return Card(
      margin: const EdgeInsets.all(0.1),
      child: ListView(
        children: [
          const SizedBox(height: 10),
          Center(
            child: Wrap(
              spacing: 10,
              runSpacing: 10,
              children: getSelectedChips(widget.selectedObjects),
            ),
          ),
          const SizedBox(height: 15),
          const Center(
            child: Text(
              "Paramètres :",
              style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(right: 60, left: 60, bottom: 8),
            child: TextField(
              onChanged: (value) => runFilter(value),
              decoration: const InputDecoration(
                  isDense: true,
                  labelText: 'Rechercher',
                  suffixIcon: Icon(Icons.search)),
            ),
          ),
          GridView.extent(
            primary: false,
            padding: const EdgeInsets.symmetric(horizontal: 50),
            // crossAxisSpacing: 10,
            // mainAxisSpacing: 10,
            maxCrossAxisExtent: 400.0,
            childAspectRatio: (1 / .12),
            shrinkWrap: true,
            children: values.keys.map((String key) {
              return CheckboxListTile(
                controlAffinity: ListTileControlAffinity.leading,
                title: Text(key),
                value: _selectedAttrs.contains(key),
                dense: true,
                onChanged: (bool? value) {
                  setState(() {
                    values[key] = value!;
                    if (value) {
                      labelRow.add(label(key));
                      actionRow.add(button(actionRow.length - 1));
                      _selectedAttrs.add(key);
                    } else {
                      labelRow.removeWhere((element) {
                        if (element.key == Key(key)) {
                          actionRow.removeAt(labelRow.indexOf(element));
                          _selectedAttrs.remove(key);
                          return true;
                        } else
                          return false;
                      });
                    }
                  });
                },
              );
            }).toList(),
          ),
          const SizedBox(
            height: 40,
          ),
          const Center(
            child: Text(
              "Actions :",
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
            ),
          ),
          Container(
            // width: 500,
            margin: const EdgeInsets.only(top: 5),
            // color: Colors.white,
            padding: const EdgeInsets.all(20.0),
            child: labelRow.length > 1
                ? Table(
                    border: TableBorder.all(
                      width: 0.05,
                    ),
                    defaultVerticalAlignment: TableCellVerticalAlignment.middle,
                    children: [
                      TableRow(children: labelRow),
                      TableRow(children: actionRow),
                    ],
                  )
                : const Center(
                    child: Text(
                      "Sélectionnez un paramètre.",
                      style:
                          TextStyle(fontSize: 13, fontWeight: FontWeight.w400),
                    ),
                  ),
          )
        ],
      ),
    );
  }

  List<Widget> getSelectedChips(List<String> nodes) {
    List<Widget> chips = [];
    nodes.forEach((value) => chips.add(RawChip(
          backgroundColor: Colors.lightGreen.shade200,
          label: Text(
            value,
            style: TextStyle(
              color: Colors.green.shade900,
              fontWeight: FontWeight.w600,
            ),
          ),
          avatar: Icon(
            Icons.check_circle,
            size: 20,
            color: Colors.green.shade900,
          ),
        )));
    return chips;
  }
}
