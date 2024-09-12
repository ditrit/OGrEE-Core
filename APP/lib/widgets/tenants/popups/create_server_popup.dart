import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/common/snackbar.dart';
import 'package:ogree_app/common/theme.dart';

class CreateServerPopup extends StatefulWidget {
  Function() parentCallback;
  CreateServerPopup({super.key, required this.parentCallback});

  @override
  State<CreateServerPopup> createState() => _CreateServerPopupState();
}

enum AuthOption { pKey, password }

class _CreateServerPopupState extends State<CreateServerPopup> {
  final _formKey = GlobalKey<FormState>();
  String? _sshHost;
  String? _sshUser;
  String? _sshKey;
  String? _sshKeyPass;
  String? _sshPassword;
  String? _installPath;
  String? _port;
  String? _kubeDns;
  bool _isLoading = false;
  AuthOption? _authOption = AuthOption.pKey;
  bool _shouldStartup = false;
  bool _isSmallDisplay = false;

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;
    _isSmallDisplay = IsSmallDisplay(MediaQuery.of(context).size.width);
    return Center(
      child: Container(
        width: 500,
        constraints: BoxConstraints(
          maxHeight: backendType == BackendType.kubernetes ? 470 : 560,
        ),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: PopupDecoration,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(40, 20, 40, 15),
          child: Form(
            key: _formKey,
            child: ScaffoldMessenger(
              child: Builder(
                builder: (context) => Scaffold(
                  backgroundColor: Colors.white,
                  body: ListView(
                    padding: EdgeInsets.zero,
                    children: [
                      Center(
                        child: Text(
                          backendType == BackendType.kubernetes
                              ? localeMsg.createKube
                              : localeMsg.createServer,
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                      ),
                      // const Divider(height: 45),
                      const SizedBox(height: 20),
                      getFormField(
                        save: (newValue) => _sshHost = newValue,
                        label: localeMsg.sshHost,
                        icon: Icons.dns,
                      ),
                      getFormField(
                        save: (newValue) => _sshUser = newValue,
                        label: localeMsg.sshUser,
                        icon: Icons.person,
                      ),
                      const SizedBox(height: 4),
                      Wrap(
                        children: <Widget>[
                          SizedBox(
                            width: 200,
                            child: RadioListTile<AuthOption>(
                              activeColor: Colors.blue.shade600,
                              dense: true,
                              title: Text(localeMsg.privateKey),
                              value: AuthOption.pKey,
                              groupValue: _authOption,
                              onChanged: (AuthOption? value) {
                                setState(() {
                                  _authOption = value;
                                });
                              },
                            ),
                          ),
                          SizedBox(
                            width: 200,
                            child: RadioListTile<AuthOption>(
                              activeColor: Colors.blue.shade600,
                              dense: true,
                              title: Text(localeMsg.password),
                              value: AuthOption.password,
                              groupValue: _authOption,
                              onChanged: (AuthOption? value) {
                                setState(() {
                                  _authOption = value;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      if (_authOption == AuthOption.pKey)
                        Column(
                          children: [
                            getFormField(
                              save: (newValue) => _sshKey = newValue,
                              label:
                                  "${localeMsg.sshPrivateKey} (/local/path/file)",
                              icon: Icons.lock,
                            ),
                            getFormField(
                              save: (newValue) => _sshKeyPass = newValue,
                              label:
                                  "${localeMsg.sshPrivateKeyPassphrase} (${localeMsg.optional})",
                              icon: Icons.lock,
                              shouldValidate: false,
                            ),
                          ],
                        )
                      else
                        getFormField(
                          save: (newValue) => _sshPassword = newValue,
                          label: localeMsg.password,
                          icon: Icons.lock,
                        ),
                      if (backendType == BackendType.kubernetes)
                        getFormField(
                          save: (newValue) => _kubeDns = newValue,
                          label: "Cluster DNS",
                          icon: Icons.dns,
                        )
                      else
                        Container(),
                      if (backendType != BackendType.kubernetes)
                        getFormField(
                          save: (newValue) => _installPath = newValue,
                          label: localeMsg.serverPath,
                          icon: Icons.folder,
                        )
                      else
                        Container(),
                      if (backendType != BackendType.kubernetes)
                        getFormField(
                          save: (newValue) => _port = newValue,
                          label: localeMsg.portServer,
                          icon: Icons.onetwothree,
                          formatters: [
                            FilteringTextInputFormatter.digitsOnly,
                          ],
                        )
                      else
                        Container(),
                      SizedBox(
                        height: backendType == BackendType.kubernetes ? 0 : 13,
                      ),
                      if (backendType != BackendType.kubernetes)
                        Row(
                          children: [
                            const SizedBox(width: 40),
                            SizedBox(
                              height: 24,
                              width: 24,
                              child: Checkbox(
                                activeColor: Colors.blue.shade600,
                                value: _shouldStartup,
                                onChanged: (bool? value) =>
                                    setState(() => _shouldStartup = value!),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              localeMsg.runAtStart,
                              style: const TextStyle(
                                fontSize: 14,
                                color: Colors.black,
                              ),
                            ),
                          ],
                        )
                      else
                        Container(),
                      const SizedBox(height: 12),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton.icon(
                            style: OutlinedButton.styleFrom(
                              foregroundColor: Colors.blue.shade900,
                            ),
                            onPressed: () => Navigator.pop(context),
                            label: Text(localeMsg.cancel),
                            icon: const Icon(
                              Icons.cancel_outlined,
                              size: 16,
                            ),
                          ),
                          const SizedBox(width: 15),
                          ElevatedButton.icon(
                            onPressed: () => submitCreateServer(localeMsg),
                            label: Text(localeMsg.create),
                            icon: _isLoading
                                ? Container(
                                    width: 24,
                                    height: 24,
                                    padding: const EdgeInsets.all(2.0),
                                    child: const CircularProgressIndicator(
                                      color: Colors.white,
                                      strokeWidth: 3,
                                    ),
                                  )
                                : const Icon(Icons.check_circle, size: 16),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  submitCreateServer(AppLocalizations localeMsg) async {
    if (_formKey.currentState!.validate()) {
      _formKey.currentState!.save();
      setState(() {
        _isLoading = true;
      });

      final Map<String, dynamic> serverInfo = <String, dynamic>{
        'host': _sshHost,
        'user': _sshUser,
      };
      if (backendType == BackendType.kubernetes) {
        serverInfo['dns'] = _kubeDns;
      } else {
        serverInfo.addAll({
          'dstpath': _installPath,
          'runport': _port,
          'startup': _shouldStartup,
        });
      }
      if (_authOption == AuthOption.pKey) {
        serverInfo.addAll({
          'pkey': _sshKey,
          'pkeypass': _sshKeyPass.toString(),
        });
      } else {
        serverInfo['password'] = _sshPassword;
      }

      final messenger = ScaffoldMessenger.of(context);
      final result = await createBackendServer(serverInfo);
      switch (result) {
        case Success():
          widget.parentCallback();
          showSnackBar(messenger, localeMsg.createOK, isSuccess: true);
          if (mounted) Navigator.of(context).pop();
        case Failure(exception: final exception):
          setState(() {
            _isLoading = false;
          });
          showSnackBar(messenger, exception.toString(), isError: true);
      }
    }
  }

  Padding getFormField({
    required Function(String?) save,
    required String label,
    required IconData icon,
    String? prefix,
    String? suffix,
    List<TextInputFormatter>? formatters,
    bool shouldValidate = true,
  }) {
    return Padding(
      padding: FormInputPadding,
      child: TextFormField(
        onSaved: (newValue) => save(newValue),
        validator: (text) {
          if (shouldValidate) {
            if (text == null || text.isEmpty) {
              return AppLocalizations.of(context)!.mandatoryField;
            }
          }
          return null;
        },
        inputFormatters: formatters,
        decoration: GetFormInputDecoration(
          _isSmallDisplay,
          label,
          prefixText: prefix,
          suffixText: suffix,
          icon: icon,
        ),
        cursorWidth: 1.3,
        style: const TextStyle(fontSize: 14),
      ),
    );
  }
}
