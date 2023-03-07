# OGrEE-APP
A Flutter application for OGrEE. It includes a frontend (ogree_app) mainly compiled as a web app and a backend (ogree_app_backend) originally developed in GO to interact with OGrEE-API but currently **deprecated**. The flutter app can interact directly with OGrEE-API

## Getting Starded: Frontend
```console
cd ogree_app
```
With Flutter, we can use the same base code to build applications for different platforms (web, Android, iOS, Windows, Linux, MacOS). To understand how it works and set up your environment, check out the [Flutter](https://docs.flutter.dev/get-started/install) documentation. 

For development, you should install the Flutter SDK and all its dependencies (depending on your OS). We recommed you use VSCode with the Flutter and Dart extensions.  

### Run web app on debug mode
Only Google Chrome can run a Flutter web app on debug mode. If the `flutter doctor` command gives you a green pass, other than directly on VSCode, you can also compile and run the web app with the following:
```console
flutter run -d chrome
```

Before running it, create a **.env** file under ogree_app/ with the URL of the target OGrEE-API:
```console
API_URL=[URL of OGrEE-API]
```

## Building and running with Docker
Our dockerfile is multi-stage: the first image install flutter and its dependencies, then compiles the web app; the second image is nginx based and runs the web server for the previously compiled app.


Before building it, add a **.env** file to the backend directory with the URL and access token to the OGrEE-API to which it should connect.

To build the Docker image, run in the root of this project:
```console
docker build . -t ogree-app
```

To run a container with the built image:
```console
docker run -p 8080:80 -d ogree-app:latest
```

If all goes well, you should be able to acess the OGrEE Web App on http://localhost:8080.

## [DEPRECATED] Getting Started: Backend
```console
cd ogree_app_backend
```
This is a simple API interfacing with OGrEE-API to make the frontend life easier, providing it with OGrEE data in the format it needs. It uses GIN, a HTTP web framework in Go. 

### Building and running
You should have Go installed We are currently using the 1.19.3 version. In the backend directory, run the following to install dependecies:
```console
go mod download
```

Before running it, you should set two environmental variables to allow the backend to connect to the OGrEE-API, a .env file can be created with the following:
```console
OGREE_URL=[URL of OGrEE-API]
OGREE_TOKEN=[authentication token for given URL]
```

Then, to compile than run:
```console
go build -o ogree_app_backend
./ogree_app_backend
```