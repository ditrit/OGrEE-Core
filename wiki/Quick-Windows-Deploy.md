# How to quickly install OGrEE-Core and deploy a tenant (Windows)
An OGrEE Windows Installer is available with each new version. With it, you can choose to install locally on your Windows machine any of the following components:
* OGrEE Admin Backend: the Go backend for the flutter app to enter "SuperAdmin" mode, that is, to create and manage tenants.
* OGrEE Admin UI: a Windows native version of the flutter APP.
* OGrEE CLI: a Windows binary of the CLI to connect and interact with an OGrEE API.
* OGrEE 3D: an Unity binary for the data center viewer provided by [OGrEE-3D](https://github.com/ditrit/OGrEE-3D) that interacts with an OGrEE CLI and API.

## Install

Launch the installer and it will guide you through the installation. Once installed, launch the OGrEE Admin Backend and Admin UI. Your backend will by default run on `http://localhost:8081`. On the OGrEE Admin UI, choose that URL as the server then enter the login credentials, by default as it follows:
* Server: `http://localhost:8081`
* User: `admin`
* Password: `Ogree@148`

## Deploy Tenant
> Docker is a prerequisite to deploy tenants.

Once logged in, click at the button `Create Tenant`. A popup should appear inviting you to insert the tenant configuration: 
1. The tenant's name and admin password (you will need it later to create database backups).
2. Which apps should the tenant have. `API` is always selected, this will create an OGrEE API docker container and MongoDB container used by the API. Select `WEB` if you would like a container to be created that will provide a web interface accessible through the browser to interact with the API: see all the objects created and manage users, for example. Select `DOC` if you wish a Swagger UI container which lets you see the API documentation in your browser.
3. Type the URL addresses for the apps with external access (API, WEB and DOC). You may also choose a logo to be displayed in the web interface. 
> Note that the first time you create a tenant it may take several minutes (up to 15 minutes). Docker images will be built and this takes time.
<p align="center">
  <img src="https://github.com/ditrit/OGrEE-Core/assets/37706737/c49b89d6-b3e1-43a0-a56a-a8276e2a345c" />
</p>