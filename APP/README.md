# OGrEE-APP
A Flutter application for OGrEE. It includes a frontend (ogree_app) mainly compiled as a web app and a backend (ogree_app_backend) only used for Super Admin mode. The flutter app can interact directly with OGrEE-API.

## Getting Starded: Frontend
```console
cd ogree_app
```
With Flutter, we can use the same base code to build applications for different platforms (web, Android, iOS, Windows, Linux, MacOS). To understand how it works and set up your environment, check out the [Flutter](https://docs.flutter.dev/get-started/install) documentation. 

For development, you should install the Flutter SDK and all its dependencies (depending on your OS). We recommed you use VSCode with the Flutter and Dart extensions.  

### Run web app on debug mode
Only Google Chrome can run a Flutter web app on debug mode. If the `flutter doctor` command gives you a green pass, other than directly on VSCode, you can also compile and run the web app in the terminal. To configure the list of possible backend URLs to which the frontend can connect (displayed as a dropdown menu in the login page), you can pass it as a environment variable:
```console
flutter run -d chrome --dart-define=BACK_URLS=http://localhost:5551,https://banana.com --dart-define=ALLOW_SET_BACK=true
```

To use the frontend with just one backend, run:
```console
flutter run -d chrome --dart-define=API_URL=http://localhost:5551
```

## Building and running with Docker
Our dockerfile is multi-stage: the first image install flutter and its dependencies, then compiles the web app; the second image is nginx based and runs the web server for the previously compiled app.

To build the Docker image to use the frontend for tenants (Super Admin), run in the root of this project:
```console
docker build . -t ogree-app --build-arg BACK_URLS=http://localhost:5551,https://banana.com --build-arg ALLOW_SET_BACK=true
```

If not, to build in normal mode, set the API URL:
```console
docker build . -t ogree-app --build-arg API_URL=http://localhost:5551

```

To run a container with the built image:
```console
docker run -p 8080:80 -d ogree-app:latest
```

If all goes well, you should be able to acess the OGrEE Web App on http://localhost:8080.

## Getting Started: Backend
```console
cd ogree_app_backend
```
This is a backend that connects to a local instance of docker to create new tenants. A new tenant consists of a docker compose deployment of 4 containers: API, DB, CLI and WebApp. Once the frontend connects to this backend, it changes its interface to tenant mode. For the backend API, it uses GIN, a HTTP web framework in Go. 

### Building and running
You should have Go installed We are currently using at least the 1.19.3 version. In the backend directory, run the following to install dependecies:
```console
go mod download
```

Then, to compile and run:
```console
go build -o ogree_app_backend
./ogree_app_backend
```

Or run directly:
```console
go run .
```