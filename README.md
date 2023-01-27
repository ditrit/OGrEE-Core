# OGrEE-APP
A Flutter application for OGrEE. It includes a frontend (ogree_app) mainly compiled as a web app and a backend (ogree_app_backend) developed in GO to interact with OGrEE-API.

## Getting Starded: Frontend
```console
cd ogree_app
```
With Flutter, we can use the same base code to build applications for different platforms (web, Android, iOS, Windows, Linux, MacOS). To understand how it works and set up your environment, check out the [Flutter](https://docs.flutter.dev/get-started/install) documentation. 

For development, you should install the Flutter SDK and all its dependencies (depending on your OS). We recommed you use VSCode with the Flutter and Dart extensions, it lets you run and debug directly in the main.dart file with special little buttons on top of the main function.  

### Run web app on debug mode
Only Google Chrome can run a Flutter web app on debug mode. If the `flutter doctor` command gives you a green pass, other than directly on VSCode, you can also compile and run the web app with the following:
```console
flutter run -d chrome
```

## Getting Started: Backend
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


## Building and running with Docker
Both frontend and backend run inside the same container. Our dockerfiles are multi-stage builds:

- Dockerfile (default): assumes the flutter frontend was already compiled by the user and the build is avaiable under ogree_app/build/web/
- Dockerfile.withcompile: compiles the frontend code under ogree_app/ and copies the result to the image.

For the both of them, the backend is compiled on a golang image base before handling the frontend. An ubuntu image is used to compile the frontend (.withcompile only) and run the applications.

The entrypoint is the script server.sh, which executes the backend in the background, lauching a server to listen on port 8080, and sets up a server with python for the frontend listening on port 5000.

Before building it, add a .env file to the backend directory with the URL and access token to the OGrEE-API to which it should connect.

To build the Docker image, run in the root of this project:
```console
docker build . -t ogree-app
```

To run a container with the built image:
```console
docker run -p 8081:5000 -p 8080:8080 -d ogree-app:latest
```

If all goes well, you should be able to acess the OGrEE Web App on http://localhost:8080.