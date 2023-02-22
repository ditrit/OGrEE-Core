import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/api.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/popup_dialog.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';
import 'package:ogree_app/widgets/select_date.dart';
import 'package:ogree_app/widgets/select_namespace.dart';

enum Steps { date, namespace, objects, result }

class SelectPage extends StatefulWidget {
  final String userEmail;
  Project? project;
  SelectPage({super.key, this.project, required this.userEmail});
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

  List<String> _selectedAttrs = [];
  List<String> get selectedAttrs => _selectedAttrs;

  @override
  void initState() {
    if (widget.project != null) {
      _selectedDate = widget.project!.dateRange;
      _selectedNamespace = widget.project!.namespace;
      _selectedAttrs = widget.project!.attributes;
      for (var obj in widget.project!.objects) {
        _selectedObjects[obj] = true;
      }
      _currentStep = Steps.result.index;
    }
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: myAppBar(context, widget.userEmail),
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
                  child: Text(_currentStep == Steps.values.first.index
                      ? 'Annuler'
                      : 'Précédent'),
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
    } else if (_currentStep == Steps.result.index) {
      Project project;
      bool isCreate = true;
      if (widget.project != null) {
        isCreate = false;
        project = widget.project!;
        project.dateRange = _selectedDate;
        project.namespace = _selectedNamespace;
        project.attributes = _selectedAttrs;
        project.objects = _selectedObjects.keys.toList();
      } else {
        project = Project(
            "",
            _selectedDate,
            _selectedNamespace,
            widget.userEmail,
            DateFormat('dd/MM/yyyy').format(DateTime.now()),
            true,
            true,
            false,
            _selectedAttrs,
            _selectedObjects.keys.toList(),
            [widget.userEmail]);
      }

      showCustomDialog(context, project, "Nommer ce projet", "Annuler",
          Icons.cancel_outlined, cancelProjectCallback, saveProjectCallback,
          isCreate: isCreate);
    } else {
      _loadObjects = _currentStep == (Steps.objects.index - 1) ? true : false;
      _currentStep < Steps.values.last.index
          ? setState(() => _currentStep += 1)
          : Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => ProjectsPage(userEmail: widget.userEmail),
              ),
            );
    }
  }

  cancel() {
    _loadObjects = _currentStep == (Steps.objects.index + 1) ? true : false;
    _currentStep > Steps.values.first.index
        ? setState(() => _currentStep -= 1)
        : Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => ProjectsPage(userEmail: widget.userEmail),
            ),
          );
  }

  saveProjectCallback(String userInput, Project project, bool isCreate) async {
    String response;
    project.name = userInput;
    if (isCreate) {
      response = await createProject(project);
    } else {
      response = await modifyProject(project);
    }
    if (response == "") {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => ProjectsPage(
            userEmail: widget.userEmail,
          ),
        ),
      );
    } else {
      showSnackBar(context, response, isError: true);
    }
  }

  cancelProjectCallback(String id) {
    Navigator.pop(context);
  }
}
