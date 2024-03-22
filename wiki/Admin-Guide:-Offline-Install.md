# How to install OGrEE in a server without internet connection 

Some servers do not have an internet connection to download this repository or get docker images and application packages from our registry online. This page helps install OGrEE in this type on environment, assuming a Linux server. As a prerequisite, the offline server must have docker installed and your personal computer or workstation must have a way to transfer packages to the offline server (ssh connection, for example). Packages needed for the server:
* Compressed files with docker images:
`ogree-api-<VERSION>.tar mongodb-6.0.9.tar ogree-webapp-<VERSION>.tar swagger-ui.tar` 
* Compressed file with OGrEE Admin Backend (Linux).

## Docker images

Transfer all the .tar files provided for the docker images to the server. In the server, to load the images, run the following command for each transferred file. 
> Note that `sudo` may be needed to run docker commands.
```
docker load -i /path/filename.tar
```

## OGrEE Admin Backend

Transfer the compressed file with the Admin Backend to the server. In the server, decompress the file, example:
```
unzip OGrEE_APP_Backend_Linux.zip -d OGrEE_APP_Backend_Linux
```
Then run the backend:
```
cd OGrEE_APP_Backend_Linux/APP/ogree_app_backend/
./ogree_app_backend
```
The backend will run on port `8081`. With an OGrEE Admin App running locally in your workstation, you can connect to that port of the server and start creating tenants.



