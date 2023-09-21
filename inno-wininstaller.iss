; Inno Setup Script

#define MyAppName "OGrEE"
#define MyBackAppName "OGrEE Admin Backend"
#define MyFrontAppName "OGrEE Admin UI"
#define MyCliAppName "OGrEE CLI"
#define My3DAppName "OGrEE 3D"
#define MyAppVersion GetEnv('VERSION')
#define MyAppPublisher "DitRit"
#define MyAppURL "https://ditrit.io/"
#define MyBackAppExeName "ogree_app_backend.exe"
#define MyFrontAppExeName "ogree_app.exe"
#define MyCliExeName "cli.exe"
#define My3DExeName "OGrEE-3D.exe"

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
Name: "full"; Description: "Full installation (APP+CLI+3D)"
Name: "custom"; Description: "Custom installation"; Flags: iscustom

[Components]
Name: "front"; Description: {#MyFrontAppName}; Types: full 
Name: "back"; Description: {#MyBackAppName}; Types: full
Name: "cli"; Description: {#MyCliAppName}; Types: full
Name: "unity"; Description: {#My3DAppName}; Types: full

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "BACK\docker-backend\{#MyBackAppExeName}"; DestDir: "{app}"; Flags: ignoreversion; Components: back
Source: "BACK\docker-backend\backend-assets\*"; DestDir: "{app}\backend-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "BACK\docker-backend\flutter-assets\*"; DestDir: "{app}\flutter-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "BACK\docker-backend\tools-assets\*"; DestDir: "{app}\tools-assets"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "deploy\*"; DestDir: "{app}\deploy"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: back
Source: "APP\docker-backend\inno.env"; DestDir: "{app}"; DestName: ".env"; Flags: ignoreversion; Components: back
Source: "APP\build\windows\runner\Release\*"; DestDir: "{app}\front"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: front
Source: "ogree-icon.ico"; DestDir: "{app}"; DestName: "ogree-icon.ico"; Flags: ignoreversion; Components: back front
Source: "CLI\other\man\*"; DestDir: "{app}\cli\other\man"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: cli
Source: "{#MyCliExeName}"; DestDir: "{app}\cli"; Flags: ignoreversion; Components: cli
Source: "OGrEE-3D_win\*"; DestDir: "{app}\3d"; Flags: ignoreversion recursesubdirs createallsubdirs; Components: unity

; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{autoprograms}\{#MyAppName}\{#MyBackAppName}"; Filename: "{app}\{#MyBackAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Components: back
Name: "{autodesktop}\{#MyBackAppName}"; Filename: "{app}\{#MyBackAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Tasks: desktopicon; Components: back
Name: "{autoprograms}\{#MyAppName}\{#MyFrontAppName}"; Filename: "{app}\front\{#MyFrontAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Components: front
Name: "{autodesktop}\{#MyFrontAppName}"; Filename: "{app}\front\{#MyFrontAppExeName}"; IconFilename: "{app}\ogree-icon.ico"; Tasks: desktopicon; Components: front
Name: "{autoprograms}\{#MyAppName}\{#MyCliAppName}"; Filename: "{app}\cli\{#MyCliExeName}"; IconFilename: "{app}\ogree-icon.ico"; Components: cli
Name: "{autodesktop}\{#MyCliAppName}"; Filename: "{app}\cli\{#MyCliExeName}"; IconFilename: "{app}\ogree-icon.ico"; Tasks: desktopicon; Components: cli
Name: "{autoprograms}\{#MyAppName}\{#My3DAppName}"; Filename: "{app}\3d\{#My3DExeName}"; Components: unity
Name: "{autodesktop}\{#My3DAppName}"; Filename: "{app}\3d\{#My3DExeName}"; Tasks: desktopicon; Components: unity

[Run]
Filename: "{app}\{#MyBackAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyBackAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: back
Filename: "{app}\front\{#MyFrontAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyFrontAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: front
Filename: "{app}\cli\{#MyCliExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyCliAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: cli
Filename: "{app}\3d\{#My3DExeName}"; Description: "{cm:LaunchProgram,{#StringChange(My3DAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent; Components: unity

[UninstallDelete]
Type: filesandordirs; Name: "{app}\3d\OGrEE-3D_Data\.ogreeCache"
Type: dirifempty; Name: "{app}\3d\OGrEE-3D_Data\"
Type: dirifempty; Name: "{app}\3d"
Type: filesandordirs; Name: "{app}\cli\log.txt"
Type: filesandordirs; Name: "{app}\cli\unitylog.txt"
Type: dirifempty; Name: "{app}\cli"