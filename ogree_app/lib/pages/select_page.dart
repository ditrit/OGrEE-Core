import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/widgets/projects/project_popup.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';
import 'package:ogree_app/widgets/select_date.dart';
import 'package:ogree_app/widgets/select_namespace.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

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
    final localeMsg = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: myAppBar(context, widget.userEmail),
      body: Center(
          child: Stepper(
        type: MediaQuery.of(context).size.width > 800
            ? StepperType.horizontal
            : StepperType.vertical,
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
                      ? localeMsg.cancel
                      : localeMsg.back),
                ),
                ElevatedButton(
                  onPressed: continued,
                  child: Text(_currentStep == Steps.values.last.index
                      ? localeMsg.save
                      : localeMsg.next),
                ),
              ],
            ),
          );
        },
        steps: <Step>[
          Step(
            title: Text(localeMsg.selectDate,
                style: const TextStyle(fontSize: 14)),
            subtitle: Text(_selectedDate),
            content: const SelectDate(),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.date.index
                ? StepState.complete
                : StepState.disabled,
          ),
          Step(
            title: Text(localeMsg.selectNamespace,
                style: const TextStyle(fontSize: 14)),
            subtitle:
                _selectedNamespace == '' ? null : Text(_selectedNamespace),
            content: const SelectNamespace(),
            isActive: _currentStep >= Steps.date.index,
            state: _currentStep >= Steps.namespace.index
                ? StepState.complete
                : StepState.disabled,
          ),
          Step(
            title: Text(localeMsg.selectObjects,
                style: const TextStyle(fontSize: 14)),
            subtitle: _selectedObjects.keys.isNotEmpty
                ? Text(localeMsg.nObjects(_selectedObjects.keys.length))
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
            title: Text(localeMsg.result, style: const TextStyle(fontSize: 14)),
            content: _currentStep == Steps.result.index
                ? SizedBox(
                    height: MediaQuery.of(context).size.height - 210,
                    child: ResultsPage(
                      selectedAttrs: _selectedAttrs,
                      selectedObjects: _selectedObjects.keys.toList(),
                      namespace: "Physical",
                    ))
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
    final localeMsg = AppLocalizations.of(context)!;
    if (_currentStep == Steps.objects.index && _selectedObjects.isEmpty) {
      // Should select at least one OBJECT before continue
      showSnackBar(context, localeMsg.atLeastOneObject, isError: true);
    } else if (_currentStep == Steps.result.index) {
      // Continue of RESULT is actually Save
      Project project;
      bool isCreate = true;
      if (widget.project != null) {
        // Editing existing project
        isCreate = false;
        project = widget.project!;
        project.dateRange = _selectedDate;
        project.namespace = _selectedNamespace;
        project.attributes = _selectedAttrs;
        project.objects = _selectedObjects.keys.toList();
      } else {
        // Saving new project
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

      showProjectDialog(
          context,
          project,
          localeMsg.nameProject,
          localeMsg.cancel,
          Icons.cancel_outlined,
          cancelProjectCallback,
          saveProjectCallback,
          isCreate: isCreate);
    } else {
      _loadObjects = _currentStep == (Steps.objects.index - 1) ? true : false;
      setState(() => _currentStep += 1);
    }
  }

  cancel() {
    _loadObjects = _currentStep == (Steps.objects.index + 1) ? true : false;
    _currentStep > Steps.values.first.index
        ? setState(() => _currentStep -= 1)
        : Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => ProjectsPage(
                  userEmail: widget.userEmail, isTenantMode: false),
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
          builder: (context) =>
              ProjectsPage(userEmail: widget.userEmail, isTenantMode: false),
        ),
      );
    } else {
      showSnackBar(context, response, isError: true);
    }
  }

  cancelProjectCallback(String? id) {
    Navigator.pop(context);
  }
}
