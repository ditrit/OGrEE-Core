# OGrEE-APP
A Flutter application for OGrEE. It includes a frontend (ogree_app) mainly compiled as a web app and a backend (ogree_app_backend) only used for Super Admin mode. The flutter app can interact directly with OGrEE-API.

## Quick deploy
To quickly deploy a frontend and backend in SuperAdmin mode, just execute the launch script appropriate to your OS. This will use docker to compile both components and to run the frontend, the backend will be run locally. 
```console
# Windows (use PowerShell)
.\launch.ps1
# Linux 
./launch.sh
```

## Frontend
```console
cd ogree_app
```
With Flutter, we can use the same base code to build applications for different platforms (web, Android, iOS, Windows, Linux, MacOS). To understand how it works and set up your environment, check out the [Flutter](https://docs.flutter.dev/get-started/install) documentation.  For docker deployment, we build and run it as a web app.

### Building and running with Docker
Our dockerfile is multi-stage: the first image install flutter and its dependencies, then compiles the web app; the second image is nginx based and runs the web server for the previously compiled app.

From the root of OGrEE-Core, run the following to build the Docker image:
```console
OGrEE-Core$ docker build -f .\APP\Dockerfile . -t ogree-app
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
Instead of interacting directly with a OGrEE-API, the App can connect to the backend avaible in this same repository to enter SuperAdmin mode. In this mode, instead of creating projects to consult an OGrEE-API database, you can create new Tenants, that is, to launch new OGrEE deployments (new OGrEE-APIs). All you have to do is connect your App to the URL of an `ogree_app_backend`. 

## Backend
```console
cd ogree_app_backend
```
This is a backend that connects to a local instance of docker to create new tenants. A new tenant consists of a docker compose deployment of up to 4 containers: API, DB, WebApp and Swagger Doc. Once the frontend connects to this backend, it changes its interface to SuperAdmin mode.  

### Building with Docker
No Go installed? No problem, docker got you covered! Run the following command to build the backend binary, according to your OS:
```console
# Windows 
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
# Linux 
docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=linux golang go build -o ogree_app_backend
# MacOS 
docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=darwin golang go build -o ogree_app_backend
```

### Building with Go
To build it, you should have Go installed (version >= 1.20). To run it, first docker should be up and running.

In the backend directory, run the following to install dependecies:
```console
go mod download
```

Then, to compile and run:
```console
go build -o ogree_app_backend
./ogree_app_backend
```

### Configuring
It is mandatory to have the `deploy` folder of OGrEE-Core to properly run the backend. A .env file should also be present under `ogree_app_backend/` with the following format:
```
TOKEN_SECRET=yoursecretstring
TOKEN_HOUR_LIFESPAN=1
ADM_PASSWORD=adminHashedPassword
DEPLOY_DIR=../../deploy/
```

Only one user (admin) can login to the superadmin backend with the password that should be added *hashed* to the .env file. If DEPLOY_DIR is omitted, the default as given in the example will be set. Example of hashed password that translates to `Ogree@148`:
```
ADM_PASSWORD="\$2a\$10\$YlOHvFzIBKzfgSxLLQkT0.7PeMsMGv/LhlL0FzDS63XKIZCCDRvim"
```

A default .env is provided in the repo with the password above.

### Running
Since the backend connects to docker to launch containers, it has to be run **locally**. To choose in what port the backend should run (default port is 8082):
```
./ogree_app_backend -port 8083
```

### Cross compile
To cross compile, that is, compile to a different OS than the one you are using:
```console
# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o ogree_app_backend_linux
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o ogree_app_backend_win
# MacOS 64-bit
GOOS=darwin GOARCH=amd64 go build -o ogree_app_backend_mac
```