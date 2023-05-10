# Flutter OGrEE-APP

An application that connects to an OGrEE-API and lets the user visualize and create reports of the complete hierarchy of objects (sites, buildings, rooms, racks, devices, etc.) and all their attributes.

Flutter allows us to compile the application to multiple target platforms. This one has been tested Web, Windows and Linux app. Check out the [flutter docs](https://docs.flutter.dev/get-started/install)  to understand how it works and install it.
## Build and run the application

Before running it, create a **.env** file in this directory with the URL of the target OGrEE-API:
```console
API_URL=[URL of OGrEE-API]
```

Executing `flutter run` in this directory will compile and run the app in debug mode on the platform locally available. If more than one available, it gives you a list of possible devices to choose from (web or windows, for example).

To compile for production, use `flutter build` followed by the target platform (`web`, for example). The result will be under /build.

## Understanding the app and the code

In Flutter, everything is a widget (visual componentes, services, models, all of it!). It all starts with `main` that calls for the creation of a `MaterialApp` widget that has a theme definition applied for the whole app (for example: default font size for titles, default button colors) and calls the home widget `LoginPage`.

`LoginPage` will call `api.dart` functions to communicate with the backend (OGrEE-API, URL from .env file). It's in that file, located in a common folder accessible by all widgets in the project, that we handle HTTP requests. 

Once successfully logged in, `ProjectsPage` is called. The API is once again called to load projects from the OGrEE-API. A project is a set of previously choosen namespace, dataset date range, objects (site, building, device, etc.) and attributes (height, weight, vendor, etc.).  Our `models/project` converts the API JSON response to valide project objects, easy to manipulate by other widgets.

The create new project button calls `SelectPage`, which creates a stepper with 4 steps that will call each a different widget as its content. The first 3 steps will allow the user to select date (`SelectDate`), namespace (`SelectNamespace`) and objects (`SelectObjects`). The latter will call the API to get the full hierarchy of objects available and display it in an expandable tree view with a settings view with filtering options to its right. 

The final step will call `ResultsPage` that displays a table with the selected objects in the first column. The next columns can be added by the user with a button that opens a popup menu with attribute options. For each selected attributes, a column is added. Math functions can be added by the user to create rows that sum or average the values in each column, if numerical. The save button will communicate with the API to save the project.

## Language translations

The app is translated in English and French, following the official flutter [guide](https://docs.flutter.dev/development/accessibility-and-localization/internationalization). Under the *l10n/* folder, one file for each language contains all the phrases used in the app. Those can be called anywhere in the app with `AppLocalizations.of(context).nameOfThePhrase`. To add a new language, a new file must be added in that folder. 
