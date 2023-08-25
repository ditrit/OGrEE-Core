; Inno Setup Script

#define MyAppName "OGrEE Admin App"
#define MyBackAppName "OGrEE Admin Backend"
#define MyFrontAppName "OGrEE Admin UI"
#define MyAppVersion GetEnv('VERSION')
#define MyAppPublisher "DitRit"
#define MyAppURL "https://ditrit.io/"
#define MyBackAppExeName "ogree_app_backend.exe"
#define MyFrontAppExeName "ogree_app.exe"

[Setup]
; NOTE: The value of AppId uniquely identifies this application. Do not use the same AppId value in installers for other applications.
; (To generate a new GUID, click Tools | Generate GUID inside the IDE.)
AppId={{5C1F8849-2EF7-459B-948A-BC328FFEAA31}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
;AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DisableProgramGroupPage=yes
; Remove the following line to run in administrative install mode (install for all users.)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
OutputBaseFilename=ogree-app-installer
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Types]
Name: "full"; Description: "Full installation (UI + Backend)"
Name: "front"; Description: "Only {#MyFrontAppName}"
Name: "back"; Description: "Only {#MyBackAppName}"

[Components]
Name: "front"; Description: "{#MyFrontAppName} Files"; Types: full front
Name: "back"; Description: "{#MyBackAppName} Files"; Types: full back

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "ogree_app_backend\{#MyBackAppExeName}"; DestDir: "{app}"; Flags: ignoreversion; Components: back
Source: "ogree_app_backend\backend-assets\*"; DestDir: "{app}\backend-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "ogree_app_backend\flutter-assets\*"; DestDir: "{app}\flutter-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "ogree_app_backend\tools-assets\*"; DestDir: "{app}\tools-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "..\deploy\*"; DestDir: "{app}\deploy"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "ogree_app_backend\inno.env"; DestDir: "{app}"; DestName: ".env"; Flags: ignoreversion; Components: back
Source: "ogree_app\build\windows\runner\Release\*"; DestDir: "{app}\front"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: front
Source: "ogree-icon.ico"; DestDir: "{app}"; DestName: "ogree-icon.ico"; Flags: ignoreversion; Components: back front
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{autoprograms}\{#MyBackAppName}"; Filename: "{app}\{#MyBackAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Components: back
Name: "{autodesktop}\{#MyBackAppName}"; Filename: "{app}\{#MyBackAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Tasks: desktopicon; Components: back
Name: "{autoprograms}\{#MyFrontAppName}"; Filename: "{app}\front\{#MyFrontAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Components: front
Name: "{autodesktop}\{#MyFrontAppName}"; Filename: "{app}\front\{#MyFrontAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Tasks: desktopicon; Components: front

[Run]
Filename: "{app}\{#MyBackAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyBackAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: back
Filename: "{app}\front\{#MyFrontAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyFrontAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: front

