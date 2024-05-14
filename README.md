<div align="center">
<img src="https://raw.githubusercontent.com/ditrit/OGrEE-Core/main/APP/assets/custom/logo.png" width="300" alt="NetBox logo" />
<p><strong>A smart datacenter digital twin</strong></p>

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ditrit_OGrEE-Core&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=ditrit_OGrEE-Core)
[![⚙️ Build - Publish](https://github.com/ditrit/OGrEE-Core/actions/workflows/build-deploy.yaml/badge.svg)](https://github.com/ditrit/OGrEE-Core/actions/workflows/build-deploy.yaml)

</div>

OGrEE exists so managing a datacenter can be easier. It aggregates data from different datacenter tools to create a digital twin easily accessible from your computer, smartphone or even VR/AR glasses. Its analysis capabilities helps identify issues faster, better prepare for maintenance and migrations, minimize risks and errors.

<p align="center">
  <a href="#why-ogree">Why OGrEE</a> |
  <a href="#ogree-core">OGrEE-Core</a> |
  <a href="#ogree-3d">OGrEE-3D</a> |
  <a href="#quickstart">Quickstart</a> |
    <a href="#quick-demo">Quick Demo</a> |
  <a href="#get-involved">Get Involved</a>
</p>

![ogree-schema](https://github.com/ditrit/OGrEE-Core/assets/37706737/78e512d0-0f24-4475-b38e-446bf3561e74)


## Why OGrEE
Different professionals work together in a datacenter with its different tools, OGrEE is here to integrate them. Here are some use cases and questions OGrEE may help with:
- An **AR/VR** headset connected to OGrEE can give an administrator or expert live access to help an in-place technician. AR/VR view can also guide a technician to find the exact server and disk that needs to be replaced.
- OGrEE's **3D view** from OGrEE can help the datacenter team organize space and prepare new installations. 
- **Data analysis** in the 2D view can let you know that those docker containers that are not performing as they should are all running in servers from the same rack that is warmer than it should. A port in the switch needs to be unplugged? A power pnale needs maintenance? OGrEE can tell you what may be impacted, from racks and servers to even software components.
- **Migrating** to cloud? Have a complete view of your datacenter and mitigate impacts with OGrEE.

OGrEE has an **offline mode**, working just with dumps and logs extracted from other tools to avoiding floating the network with constant requests.

## OGrEE-Core
OGrEE-Core assembles 3 essential components of OGrEE:
- **API**: an API developed in Go with a MongoDB to store all the datacenter information (sites, buildings, devices, applications; etc.) and provide secure access to it.
- **APP**: APP is a Flutter application that can run as a native app (iOS, Android, Windows, Linux, Mac) or webapp (all main browsers) and its Go backend to view and interact with the datacenter data, providing reports and analysis.
- **CLI**: command line client to interact with the data from the API and to pilot the 3D view from OGrEE-3D.

Together, these components form an **OGrEE Tenant**, a deployment instance of OGrEE. 
<div align="center">
    
![ogree-schema](https://github.com/ditrit/OGrEE-Core/assets/37706737/378c6cbe-aea2-4db0-82d6-6c3a18ecc6c5)

</div>

## OGrEE-3D
This is OGrEE's 3D client, a 3D datacenter viewer based on Unity game engine to display and interact with an OGrEE Tenant.
You can access to the OGrEE-3D repo [here](https://github.com/ditrit/OGrEE-3D).  

## Quickstart

A few options are avaiable to help deploy your first OGrEE Tenant:

#### Option 1: SuperAdmin APP 
The APP has a **SuperAdmin** version to create and manage OGrEE Tenants with a pretty UI. To quickly deploy it, just execute the launch script appropriate to your OS from the `deploy/app` folder. 
```console
cd deploy/app

# Windows (use PowerShell)
.\launch.ps1
# Linux 
./launch.sh
# MacOS 
./launch.sh -m
```
> This will use docker to compile APP and BACK (a Go backend for SuperAdmin), then run a docker container for the SuperAdmin webapp and locally run the compiled binary of BACK. For more launch options, check the documentation under [deploy](https://github.com/ditrit/OGrEE-Core/tree/main/deploy).

Check the [SuperAdmin user guide](https://github.com/ditrit/OGrEE-Core/wiki/Quick-Windows-Deploy) on how to create and manage a tenant as well as download a CLi and 3D client from it.

#### Option 2: Docker compose
Don't want a pretty UI to manage a tenant? Just run docker compose to create a new tenant:

```docker compose --project-name <your-project> --profile web -f deploy/docker/docker-compose.yml up```
> This will create a docker deployment with an API, a DB and a WebAPP

The config can be updated beforehand in ```deploy/docker/.env```

#### Option 3: Windows Installer (Windows only)
We have a Windows Installer to quickly install SuperAdmin with a CLI and a 3D client. Download the latest from here and it will guide through the installation. Then, check the [SuperAdmin user guide](https://github.com/ditrit/OGrEE-Core/wiki/Quick-Windows-Deploy) on how to create and manage a tenant

## OGrEE-Tools
OGrEE-Tools is a collection of tools help populate OGrEE with data. They can help extract and parse data from multiple logs, create 3D models of servers using Machine Learning and much more. Check out its repo [here](https://github.com/ditrit/OGrEE-Tools). 

## Quick Demo

https://github.com/ditrit/OGrEE-Core/assets/35805113/d6fe1a3e-9c5f-42e7-b926-1c6211a7df0d

## Get Involved

New contributors are more than welcome!
- Have an idea? Let's talk about it on Discord or on the [discussion forum](https://github.com/ditrit/OGrEE-Core/discussions).
- Want to to code? Check out our [how to contribute](https://github.com/ditrit/OGrEE-Core/wiki/How-to-contribute-(Dev-Guide)) guide. 


