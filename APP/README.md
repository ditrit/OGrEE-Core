# OGrEE-APP
A Flutter application for OGrEE. This frontend can interact with the backends of BACK folder, entering SuperAdmin mode, and also with the OGrEE-API of API folder.

## Quick deploy
To quickly deploy a frontend and docker backend in SuperAdmin mode, just execute the launch script appropriate to your OS from the `deploy/app` folder. This will use docker to compile both components and to run the frontend, the backend will be run locally. 
```console
# Windows (use PowerShell)
.\launch.ps1
# Linux 
./launch.sh
# MacOS 
./launch.sh -m
```
For more options, check the documentation under `deploy/`.

## Flutter Frontend

An application that connects to an OGrEE-API and lets the user visualize and create reports of the complete hierarchy of objects (sites, buildings, rooms, racks, devices, etc.) and all their attributes. The app can also connect to a backend (BACK folder), entering SuperAdmin mode, where it can be used to launch and manage tenants (deployments of OGrEE).

With Flutter, we can use the same base code to build applications for different platforms (web, Android, iOS, Windows, Linux, MacOS). To understand how it works and set up your environment, check out the [Flutter](https://docs.flutter.dev/get-started/install) documentation.  

## Building and running with Docker
For docker deployment, we build and run it as a web app.
Our dockerfile is multi-stage: the first image install flutter and its dependencies, then compiles the web app; the second image is nginx based and runs the web server for the previously compiled app.

From the root of OGrEE-Core, run the following to build the Docker image:
```console
OGrEE-Core$ docker build . -f APP/Dockerfile -t ogree-app
```

To run a container with the built image and expose the app in the local port 8080:
```console
docker run -p 8080:80 -d ogree-app:latest
```

If all goes well, you should be able to acess the OGrEE Web App on http://localhost:8080.

### Pick which OGrEE-API to connect
You can configure to which API you wish to connect. This is set by a `.env` file located under `ogree_app/assets/custom` that should contain the following definitions:
```
API_URL=http://localhost:3001
ALLOW_SET_BACK=false
BACK_URLS=http://localhost:3001,http://localhost:8082
```

- If `ALLOW_SET_BACK=true`, the App will display a dropdown menu in the login page allowing the user to type the URL of the API to connect. `BACK_URLS` will be the selectable hints displayed when the use click on the dropdown menu, serving as shortcuts.
- If `ALLOW_SET_BACK=false`, the App will only connect to the given `API_URL` and not give the user any choice. Instead of the dropdown menu, a logo will be displayed, the image file at: `ogree_app/assets/custom/logo.png`.

### Pick which OGrEE-API under docker
The easiest way to edit the `.env` file in a docker container is to mount your local folder containing the file and a logo image (not mandatory) as a volume when running it:
```
docker run -p 8080:80 -v [your/custom/folder]:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest
```

### Frontend SuperAdmin mode
Instead of interacting directly with a OGrEE-API, the App can connect to the backend available in this same repository (BACK folder) to enter SuperAdmin mode. In this mode, instead of creating projects to consult an OGrEE-API database, you can create new Tenants, that is, to launch new OGrEE deployments (new OGrEE-APIs). All you have to do is connect your App to the URL of an `ogree_app_backend`. 

## Developing
For development, you should install the Flutter SDK and all its dependencies (depending on your OS). We recommed you use VSCode with the Flutter and Dart extensions. 

### Language translations

The app is translated in English and French, following the official flutter [guide](https://docs.flutter.dev/development/accessibility-and-localization/internationalization). Under the *l10n/* folder, one file for each language contains all the phrases used in the app. Those can be called anywhere in the app with `AppLocalizations.of(context).nameOfThePhrase`. To add a new language, a new file must be added in that folder. 

### Understanding the app and the code (regular mode)

In Flutter, everything is a widget (visual componentes, services, models, all of it!). It all starts with `main` that calls for the creation of a `MaterialApp` widget that has a theme definition applied for the whole app (for example: default font size for titles, default button colors) and calls the home widget `LoginPage`.

`LoginPage` will call `api.dart` functions to communicate with the backend through HTTP requests. Once successfully logged in, `ProjectsPage` is called. The API is once again called to load projects from the OGrEE-API or tenants from the backend if in SuperAdmin mode. 

In normal mode, a project is a set of previously choosen namespace, dataset date range, objects (site, building, device, etc.) and attributes (height, weight, vendor, etc.).  Our `models/project` converts the API JSON response to valide project objects, easy to manipulate by other widgets.

The create new project button calls `SelectPage`, which creates a stepper with 4 steps that will call each a different widget as its content. The first 3 steps will allow the user to select date (`SelectDate`), namespace (`SelectNamespace`) and objects (`SelectObjects`). The latter will call the API to get the full hierarchy of objects available and display it in an expandable tree view with a settings view with filtering options to its right. 

The final step will call `ResultsPage` that displays a table with the selected objects in the first column. The next columns can be added by the user with a button that opens a popup menu with attribute options. For each selected attributes, a column is added. Math functions can be added by the user to create rows that sum or average the values in each column, if numerical. The save button will communicate with the API to save the project.

