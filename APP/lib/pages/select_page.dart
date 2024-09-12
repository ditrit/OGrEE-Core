import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/pages/impact_page.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:ogree_app/pages/results_page.dart';
import 'package:ogree_app/widgets/projects/project_popup.dart';
import 'package:ogree_app/widgets/select_date.dart';
import 'package:ogree_app/widgets/select_namespace.dart';
import 'package:ogree_app/widgets/select_objects/select_objects.dart';

enum Steps { date, namespace, objects, result }

class SelectPage extends StatefulWidget {
  final String userEmail;
  Project? project;
  bool isImpact;
  SelectPage({
    super.key,
    this.project,
    required this.userEmail,
    this.isImpact = false,
  });
  @override
  State<SelectPage> createState() => SelectPageState();

  static SelectPageState? of(BuildContext context) =>
      context.findAncestorStateOfType<SelectPageState>();
}

class SelectPageState extends State<SelectPage> with TickerProviderStateMixin {
  // Flow control
  Steps _currentStep = Steps.date;
  bool _loadObjects = false;

  // Shared data used by children in stepper
  String _selectedDate = '';
  set selectedDate(String value) => _selectedDate = value;

  Namespace _selectedNamespace = Namespace.Test;
  set selectedNamespace(Namespace value) => _selectedNamespace = value;

  Map<String, bool> _selectedObjects = {};
  Map<String, bool> get selectedObjects => _selectedObjects;
  set selectedObjects(Map<String, bool> value) => () {
        _selectedObjects = value;
      };

  List<String> _selectedAttrs = [];
  List<String> get selectedAttrs => _selectedAttrs;
  set selectedAttrs(List<String> value) => () {
        _selectedAttrs = value;
      };

  @override
  void initState() {
    if (widget.project != null) {
      // select date and namespace from project
      _selectedDate = widget.project!.dateRange;
      _selectedNamespace = Namespace.values.firstWhere(
        (e) => e.toString() == 'Namespace.${widget.project!.namespace}',
      );
      _selectedAttrs = widget.project!.attributes;
      // select objects
      for (final obj in widget.project!.objects) {
        _selectedObjects[obj] = true;
      }
      // adjust step
      widget.isImpact = widget.project!.isImpact;
      if (widget.project!.lastUpdate == "AUTO") {
        // auto project
        _loadObjects = true;
        _currentStep = Steps.objects;
        widget.project = null;
      } else {
        _currentStep = Steps.result;
      }
    } else if (widget.isImpact) {
      _selectedNamespace = Namespace.Physical;
      _loadObjects = true;
      _currentStep = Steps.objects;
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
          type: MediaQuery.of(context).size.width > 650
              ? StepperType.horizontal
              : StepperType.vertical,
          physics: const ScrollPhysics(),
          currentStep: _currentStep.index,
          // onStepTapped: (step) => tapped(step),
          controlsBuilder: (context, _) {
            return Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  TextButton(
                    onPressed: cancel,
                    child: Text(
                      _currentStep == Steps.values.first
                          ? localeMsg.cancel
                          : localeMsg.back,
                    ),
                  ),
                  ElevatedButton(
                    onPressed: continued,
                    child: Text(
                      _currentStep == Steps.values.last
                          ? localeMsg.save
                          : localeMsg.next,
                    ),
                  ),
                ],
              ),
            );
          },
          steps: <Step>[
            Step(
              title: Text(
                localeMsg.selectDate,
                style: const TextStyle(fontSize: 14),
              ),
              subtitle: _selectedDate != ""
                  ? Text(_selectedDate)
                  : const Icon(Icons.all_inclusive, size: 15),
              content: const SelectDate(),
              isActive: _currentStep.index >= Steps.date.index,
              state: _currentStep.index >= Steps.date.index
                  ? StepState.complete
                  : StepState.disabled,
            ),
            Step(
              title: Text(
                localeMsg.selectNamespace,
                style: const TextStyle(fontSize: 14),
              ),
              subtitle: _selectedNamespace == Namespace.Test
                  ? null
                  : Text(_selectedNamespace.name),
              content: const SelectNamespace(),
              isActive: _currentStep.index >= Steps.date.index,
              state: _currentStep.index >= Steps.namespace.index
                  ? StepState.complete
                  : StepState.disabled,
            ),
            Step(
              title: Text(
                localeMsg.selectObjects,
                style: const TextStyle(fontSize: 14),
              ),
              subtitle: _selectedObjects.keys.isNotEmpty
                  ? Text(localeMsg.nObjects(_selectedObjects.keys.length))
                  : null,
              content: SizedBox(
                height: MediaQuery.of(context).size.height > 205
                    ? MediaQuery.of(context).size.height - 220
                    : MediaQuery.of(context).size.height,
                child: SelectObjects(
                  dateRange: _selectedDate,
                  namespace: _selectedNamespace,
                  load: _loadObjects,
                ),
              ),
              isActive: _currentStep.index >= Steps.date.index,
              state: _currentStep.index >= Steps.objects.index
                  ? StepState.complete
                  : StepState.disabled,
            ),
            Step(
              title:
                  Text(localeMsg.result, style: const TextStyle(fontSize: 14)),
              content: _currentStep == Steps.result
                  ? (widget.isImpact
                      ? ImpactPage(
                          selectedObjects: _selectedObjects.keys.toList(),
                        )
                      : ResultsPage(
                          dateRange: _selectedDate,
                          selectedAttrs: _selectedAttrs,
                          selectedObjects: _selectedObjects.keys.toList(),
                          namespace: _selectedNamespace.name,
                        ))
                  : const Center(child: CircularProgressIndicator()),
              isActive: _currentStep.index >= Steps.date.index,
              state: _currentStep.index >= Steps.result.index
                  ? StepState.complete
                  : StepState.disabled,
            ),
          ],
        ),
      ),
    );
  }

  continued() async {
    final localeMsg = AppLocalizations.of(context)!;
    _loadObjects = false;
    switch (_currentStep) {
      case (Steps.date):
        setState(() => _currentStep = Steps.namespace);
      case (Steps.namespace):
        _loadObjects = true;
        _selectedObjects = {};
        setState(() => _currentStep = Steps.objects);
      case (Steps.objects):
        if (_selectedObjects.isEmpty) {
          // Should select at least one OBJECT before continue
          showSnackBar(
            ScaffoldMessenger.of(context),
            localeMsg.atLeastOneObject,
            isError: true,
          );
        } else {
          _selectedAttrs = [];
          setState(() => _currentStep = Steps.result);
        }
      case (Steps.result):
        // Continue of RESULT is actually Save
        Project project;
        bool isCreate = true;
        if (widget.project != null && widget.project!.lastUpdate != "AUTO") {
          // Editing existing project
          isCreate = false;
          project = widget.project!;
          project.dateRange = _selectedDate;
          project.namespace = _selectedNamespace.name;
          project.attributes = _selectedAttrs;
          project.objects = _selectedObjects.keys.toList();
        } else {
          // Saving new project
          project = Project(
            "",
            _selectedDate,
            _selectedNamespace.name,
            widget.userEmail,
            DateFormat('dd/MM/yyyy').format(DateTime.now()),
            true,
            true,
            false,
            _selectedAttrs,
            _selectedObjects.keys.toList(),
            [widget.userEmail],
            isImpact: widget.isImpact,
          );
        }

        showProjectDialog(
          context,
          project,
          localeMsg.nameProject,
          saveProjectCallback,
          isCreate: isCreate,
        );
    }
  }

  void cancel() {
    _loadObjects = false;
    switch (_currentStep) {
      case (Steps.date):
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) =>
                ProjectsPage(userEmail: widget.userEmail, isTenantMode: false),
          ),
        );
      case (Steps.namespace):
        setState(() => _currentStep = Steps.date);
      case (Steps.objects):
        if (widget.isImpact) {
          showSnackBar(
            ScaffoldMessenger.of(context),
            AppLocalizations.of(context)!.onlyPredefinedWarning,
          );
          return;
        }
        setState(() => _currentStep = Steps.namespace);
      case (Steps.result):
        _loadObjects = true;
        setState(() => _currentStep = Steps.objects);
    }
  }

  saveProjectCallback(
    String userInput,
    Project project,
    bool isCreate,
    Function? callback,
  ) async {
    Result result;
    project.name = userInput;
    final messenger = ScaffoldMessenger.of(context);
    final navigator = Navigator.of(context);
    if (isCreate) {
      result = await createProject(project);
    } else {
      result = await modifyProject(project);
    }
    switch (result) {
      case Success():
        navigator.push(
          MaterialPageRoute(
            builder: (context) =>
                ProjectsPage(userEmail: widget.userEmail, isTenantMode: false),
          ),
        );
      case Failure(exception: final exception):
        showSnackBar(messenger, exception.toString(), isError: true);
    }
  }
}
