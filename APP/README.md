# OGrEE-APP
A Flutter application for OGrEE. It includes a frontend (ogree_app) mainly compiled as a web app and a backend (ogree_app_backend) only used for Super Admin mode. The flutter app can interact directly with OGrEE-API.

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

### Building and running
Since the backend connects to docker to launch containers, it has to be run **locally**. To build it, you should have Go installed. We are currently using the 1.19.3 version. To run it, first docker should be up and running.

In the backend directory, run the following to install dependecies:
```console
go mod download
```

It is mandatory to have the `deploy` folder of OGrEE-Core to properly run the backend and also a .env file which should include:
```
TOKEN_SECRET=yoursecretstring
TOKEN_HOUR_LIFESPAN=1
ADM_PASSWORD=adminHashedPassword
DEPLOY_DIR = ../../deploy
```

Only one user (admin) can login to the superadmin backend with the password that should be added *hashed* to the .env file. If DEPLOY_DIR is omitted, the default as given in the example will be set. Example of hashed password that translates to `Ogree@148`:
```
ADM_PASSWORD="\$2a\$10\$YlOHvFzIBKzfgSxLLQkT0.7PeMsMGv/LhlL0FzDS63XKIZCCDRvim"
```

Then, to compile and run:
```console
go build -o ogree_app_backend
./ogree_app_backend
```

To choose in what port the backend should run (default port is 8082):
```
./ogree_app_backend -port 8083
```

To cross compile:
```console
# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o ogree_app_backend_linux
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o ogree_app_backend_win
# MacOS 64-bit
GOOS=darwin GOARCH=amd64 go build -o ogree_app_backend_mac
```