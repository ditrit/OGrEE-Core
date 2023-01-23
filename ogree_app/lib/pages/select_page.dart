import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';
import 'package:ogree_app/widgets/select_date.dart';
import 'package:ogree_app/widgets/select_namespace.dart';

enum Steps { date, namespace, objects, result }

class SelectPage extends StatefulWidget {
  const SelectPage({super.key});
  @override
  State<SelectPage> createState() => _SelectPageState();

  static _SelectPageState? of(BuildContext context) =>
      context.findAncestorStateOfType<_SelectPageState>();
}

class _SelectPageState extends State<SelectPage> with TickerProviderStateMixin {
  // Flow control
  int _currentStep = 0;
  bool _loadObjects = false;

  // Shared data used by children in stepper
  String _selectedDate = DateFormat('dd/MM/yyyy').format(DateTime.now());
  set selectedDate(String value) => _selectedDate = value;

  String _selectedNamespace = '';
  set selectedNamespace(String value) => _selectedNamespace = value;

  Map<String, bool> _selectedObjects = {};
  Map<String, bool> get selectedObjects => _selectedObjects;
  set selectedObjects(Map<String, bool> value) => () {
        _selectedObjects = value;
      };

  final List<String> _selectedAttrs = [];
  List<String> get selectedAttrs => _selectedAttrs;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: myAppBar(),
      body: Center(
          child: Stepper(
        type: StepperType.horizontal,
        physics: const ScrollPhysics(),
        currentStep: _currentStep,
        // onStepTapped: (step) => tapped(step),
        controlsBuilder: (context, _) {
          return Padding(
            padding: const EdgeInsets.only(top: 10),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                TextButton(
                  onPressed: cancel,
                  child: const Text('Annuler'),
                ),
                ElevatedButton(
                  onPressed: continued,
                  child: Text(_currentStep == Steps.values.last.index
                      ? 'Sauvegarder'
                      : 'Suivant'),
                ),
              ],
            ),
          );
        },
        steps: <Step>[
          Step(
            title: const Text('Choisir date'),
            subtitle: Text(_selectedDate),
            content: const SelectDate(),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.date.index
                ? StepState.complete
                : StepState.disabled,
          ),
          Step(
            title: const Text('Choisir namespace'),
            subtitle:
                _selectedNamespace == '' ? null : Text(_selectedNamespace),
            content: const SelectNamespace(),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.namespace.index
                ? StepState.complete
                : StepState.disabled,
          ),
          Step(
            title: const Text('Choisir les objets'),
            subtitle: _selectedObjects.keys.isNotEmpty
                ? Text(_selectedObjects.keys.length == 1
                    ? '1 objet'
                    : '${_selectedObjects.keys.length} objets')
                : null,
            content: SizedBox(
                height: MediaQuery.of(context).size.height - 205,
                child: SelectObjects(
                    namespace: _selectedNamespace, load: _loadObjects)),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.objects.index
                ? StepState.complete
                : StepState.disabled,
          ),
          // Step(
          //   title: const Text('Créer la requête'),
          //   subtitle: _selectedAttrs.isNotEmpty
          //       ? Text(_selectedAttrs.length == 1
          //           ? '1 paramètre'
          //           : '${_selectedAttrs.length} paramètres')
          //       : null,
          //   content: _currentStep == Steps.attrs.index
          //       ? SizedBox(
          //           height: MediaQuery.of(context).size.height - 205,
          //           child: SelectAttributes(
          //               selectedObjects: _selectedObjects.keys.toList()))
          //       : const Center(child: CircularProgressIndicator()),
          //   isActive: _currentStep >= Steps.date.index,
          //   state: _currentStep >= Steps.attrs.index
          //       ? StepState.complete
          //       : StepState.disabled,
          // ),
          // Step(
          //   title: const Text('Résultat'),
          //   content: _currentStep == Steps.result.index
          //       ? SizedBox(
          //           height: MediaQuery.of(context).size.height - 210,
          //           child: ResultsPage(
          //               selectedAttrs: _selectedAttrs,
          //               selectedObjects: _selectedObjects.keys.toList()))
          //       : const Center(child: CircularProgressIndicator()),
          //   isActive: _currentStep >= Steps.date.index,
          //   state: _currentStep >= Steps.result.index
          //       ? StepState.complete
          //       : StepState.disabled,
          // ),
          Step(
            title: const Text('Résultat'),
            content: _currentStep == Steps.result.index
                ? SizedBox(
                    height: MediaQuery.of(context).size.height - 210,
                    child: ResultsPage(
                        selectedAttrs: _selectedAttrs,
                        selectedObjects: _selectedObjects.keys.toList()))
                : const Center(child: CircularProgressIndicator()),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.result.index
                ? StepState.complete
                : StepState.disabled,
          ),
        ],
      )),
    );
  }

  tapped(int step) {
    setState(() => _currentStep = step);
  }

  continued() {
    if (_currentStep == Steps.objects.index && _selectedObjects.isEmpty) {
      showSnackBar(context, "Sélectionnez au moins 1 objet avant d'avancer",
          isError: true);
      return;
    }
    _loadObjects = _currentStep == (Steps.objects.index - 1) ? true : false;
    _currentStep < Steps.values.last.index
        ? setState(() => _currentStep += 1)
        : Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => const ProjectsPage(),
            ),
          );
  }

  cancel() {
    _loadObjects = _currentStep == (Steps.objects.index + 1) ? true : false;
    _currentStep > Steps.values.first.index
        ? setState(() => _currentStep -= 1)
        : Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => const ProjectsPage(),
            ),
          );
  }
}
