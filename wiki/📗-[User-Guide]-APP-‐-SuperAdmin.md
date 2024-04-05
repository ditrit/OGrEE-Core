The APP is a Flutter application that can be run as a webapp or native Windows/Linux/macOS/iOS/Android app. Its SuperAdmin mode is used for creating and managing tenant and can help you get started with installing OGrEE-Core.

# How to install

### Install OGrEE-Core on Windows
An OGrEE Windows Installer is available with each new version. With it, you can choose to install locally on your Windows machine any of the following components:
* OGrEE Admin Backend: the Go backend for the flutter app to enter "SuperAdmin" mode, that is, to create and manage tenants.
* OGrEE Admin UI: a Windows native version of the flutter APP.
* OGrEE CLI: a Windows binary of the CLI to connect and interact with an OGrEE API.
* OGrEE 3D: an Unity binary for the data center viewer provided by [OGrEE-3D](https://github.com/ditrit/OGrEE-3D) that interacts with an OGrEE CLI and API.

Just download it from the [Releases](https://github.com/ditrit/OGrEE-Core/releases) page (the file with `.exe` extension) and let it guide you through the installation. Once installed, launch the OGrEE Admin Backend and Admin UI.

### Install OGrEE-Core on Linux or MacOS
To quickly deploy an APP and docker backend in SuperAdmin mode, clone the repo and execute the launch script appropriate to your OS from the `deploy/app` folder. This will use docker to compile both components and to run the frontend, the backend will be run locally.

```console
# Linux 
./launch.sh
# MacOS 
./launch.sh -m
```

Once both are running, you will be able to use the SuperAdmin APP to download a CLI and a 3D client under the button `Tools`, completing your OGrEE-Core installation.

# How to use

## Login 
Your backend will by default run on `http://localhost:8081`. On the OGrEE Admin UI, choose that URL as the server then enter the login credentials, by default as it follows:
* Server: `http://localhost:8081`
* User: `admin`
* Password: `admin`

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

## Manage Tenant
Once a tenant is created, it will be displayed as card in the main page under `Applications`.
 <p align="center">
<img src="https://github.com/ditrit/OGrEE-Core/assets/37706737/39f7a189-3145-4958-9f46-f977563cfccb" />
</p>

The ⏯️ play button starts the tenant by running a docker command to start all its containers. The ⏹️ stop button is to stop all containers.
A colored dot right next to the name hints on the tenant's status:  
🔴 means the tenant is stopped, that is, all containers from this tenant deployment are stopped.  
🟠 means that one or more containers are not running for this tenant.  
🟢 means all is good! All containers are properlly running. 

The ✏️ edit button is to modify the deployment, allowing to add or remove applications (APP, DOC), change its URL and ports, update the tenant's version. The 🗑️ remove button is for deleting the tenant with its all data (be careful!).

The 🔍 info button opens the tenant view. In the tab "Deployment", you can see more information regarding its applications (containers) status and configuration as well as check its logs. In the other tabs, you can access the API associated with this tenant (a login with the API's credentials will be requested), allowing to create and manage users, domains and tags for the tenant. 
 
## Deploy Tools
In SuperAdmin's page, you will find a Tools button with a dropdown menu giving the option to create 3 tools: 
- [Netbox](https://github.com/netbox-community/netbox)
- [Nautobot](https://github.com/nautobot/nautobot) 
- [OpenDCIM](https://github.com/opendcim/openDCIM)

Only one deployment of each tool can be created and managed by a SuperAdmin. For each tool, a creation popup will ask for the ports that should be used by the tool. The tools will be deployed as a docker deployment with multiples containers.

