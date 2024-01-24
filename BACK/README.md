# Backend
BACK contains a backend application for deploying and managing tenants. The frontend APP uses this backend to enter "SuperAdmin" mode. The backend has two modes of operation: docker or kubernetes. 

The docker mode that connects to a local instance of docker to create new tenants. A new tenant consists of a docker compose deployment of up to 4 containers: API, DB, WebApp and Swagger Doc. Once the frontend connects to this backend, it changes its interface to SuperAdmin mode.  

The kubernetes mode works in a quite similar way, connecting to a kubernetes cluster and creating namespaces for each new tenant.

## Building with Docker
No Go installed? No problem, docker got you covered! Run the following command to build the backend binary, according to your OS:
```console
# Windows 
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
# Linux 
docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=linux golang go build -o ogree_app_backend
# MacOS 
docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=darwin golang go build -o ogree_app_backend
```

## Building with Go
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

## Configuring
It is mandatory to have the `deploy` folder of OGrEE-Core to properly run the backend. A .env file should also be present under `ogree_app_backend/` with the following format:
```
TOKEN_SECRET=yoursecretstring
TOKEN_HOUR_LIFESPAN=1
ADM_PASSWORD=adminHashedPassword
DEPLOY_DIR=../../deploy/
```

Only one user (admin) can login to the superadmin backend with the password that should be added *hashed* to the .env file. If DEPLOY_DIR is omitted, the default as given in the example will be set. Example of hashed password that translates to `admin`:
```
ADM_PASSWORD="\$2a\$12\$mgyVEGO1SZmCq1Bml8V5VePzcWLnC0hbGuHa/irKgbqoLVwEL6Vb2"
```

A default .env is provided in the repo with the password above.

## Running
For docker mode, since the backend connects to docker to launch containers, it has to be run **locally**. To choose in what port the backend should run (default port is 8082):
```
./ogree_app_backend -port 8083
```

To use kubernetes mode instead, simply add the kube flag:
```
./ogree_app_backend -kube
```

## Cross compile
To cross compile from Linux or Mac (that is, compile to a different OS than the one in use), use the commands bellow. For Windows, user `set` for GOOS and GOARCH before running the go build command.
```console
# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o ogree_app_backend_linux
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o ogree_app_backend_win
# MacOS 64-bit
GOOS=darwin GOARCH=amd64 go build -o ogree_app_backend_mac
```
