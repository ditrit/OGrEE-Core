# OGrEE-Core

OGrEE-Core assembles 3 essential components of OGrEE, allowing you to create an OGrEE Tenant to store and interact with your datacenter data.

## Quick Intro
![ogree-schema](https://github.com/ditrit/OGrEE-Core/assets/37706737/378c6cbe-aea2-4db0-82d6-6c3a18ecc6c5)

An OGrEE Tenant consists of a DB populated with objects of a datacenter (sites, buildings, devices, etc.) that can be accessed through an API. For a user friendly access, a WebAPP can be deployed for each Tenant or a locally installed CLI can be used. Check the [OGrEE-3D](https://github.com/ditrit/OGrEE-3D) repo for the 3D datacenter viewer. To launch and manage a tenant, a WebAPP in "SuperAdmin" version with its backend in Go are available.

## How to deploy an OGrEE Tenant
The prefered way to deploy the API is to use the superadmin interface. See the [OGrEE-APP documentation](https://github.com/ditrit/OGrEE-Core/tree/main/APP).

## Quickstart to deploy an OGrEE Tenant without OGrEE-APP

Run:
```docker compose --project-name <your-project> -f deploy/docker/docker-compose.yml up```

The config can be updated beforehand in ```deploy/docker/.env```

## Frontend config
To use the frontend (CLI, 3D, APP), a ```config.toml``` file must be created at the root of the repo.

Example :
```
[OGrEE-CLI]
Verbose = "ERROR"
APIURL = "http://127.0.0.1:3001"
UnityURL = "127.0.0.1:5500"
UnityTimeout = "10ms"
HistPath = "./.history"
Script = ""
Drawable = ["all"]
DrawLimit = 100
Updates = ["all"]
User = ""
Variables = [
    {Name = "ROOT", Value = "path_to_root"},
    {Name = "ROOT2", Value = "$ROOT/path_to_root2"},
]

[OGrEE-CLI.DrawableJson]
tenant = "./other/drawTemplates/tenant.json"

[OGrEE-3D]
verbose = true
fullscreen = false
cachePath = "C:/"
cacheLimitMo = 100
# Port used to receive messages     from OGrEE-CLI. 5500 by default
cliPort = 5500
# Value clamped from 0 to 100
alphaOnInteract = 50

[OGrEE-3D.textures]
# Textures loaded in the 3D client at startup, in "name = url" format
perf22 = "https://raw.githubusercontent.com/ditrit/OGREE-3D/master/Assets/Resources/Textures/TilePerf22.png"
perf29 = "https://raw.githubusercontent.com/ditrit/OGREE-3D/master/Assets/Resources/Textures/TilePerf29.png"

[OGrEE-3D.colors]
# Colors can be defines with hexadecimal codes or html colors
selection = "#21FF00"
edit = "#C900FF"
focus = "#FF9F00"
highlight = "#00D5FF"

usableZone = "#DBEDF2"
reservedZone = "#F2F2F2"
technicalZone = "#EBF2DE"

[OGrEE-3D.temperature]
# Minimum and maximum values for temperatures in Celsius and Fahrenheit to define the range of the temperature gradients
minC = 0
maxC = 100
minF = 32
maxF = 212

# Define a custom gradient by defining up to 8 colors in rgba format (rgb from 0 to 255, a from 0 to 100)
useCustomGradient = true
customTemperatureGradient = [
    [0,0,255,0],
    [255,0,0,100],
    [255,255,0,50]
]
```
## How to checkout a single component
OGrEE-Core is a monorepo hosting 3 differents applications: API, CLI and APP. With git sparse checkout functionality, you can choose which component you want to checkout instead of the whole repo. 

```
mkdir sparse
cd sparse
git init
git remote add -f origin https://github.com/ditrit/OGrEE-Core.git
git sparse-checkout init
# subfolders to checkout
git sparse-checkout set "deploy" "APP"
# branch you wish to checkout
git pull origin main 
```


# 🔁 OGree-Core GitFlow

![Workflows diagram](/assets/images/main.jpg)

## Release candidate

After merging a dev branch on main, workflows will create a new branch name `release-candidate/x.x.x`.

Semver bump are define by the following rules:
- One commit between last tag and main contains: break/breaking -> Bump major version
- One commit between last tag and main contains: feat/features -> Bump minor version
- Any other cases -> Bump patch version

if a branch release-candidate with the same semver already exists, it will be deleted and recreated from the new commit.

Example: A patch is merged after another, which has not yet been released

## Release

After validate a release candidate, a manual workflow named `📦 Create Release` can be called form github actions panel on the release-candidate branch and will create a `release/x.x.x branch`

![Github Actions panel](/assets/images/github.png)

Note: If release workflow is launch on another branch other than a release-candidate, it will fail.

## Build docker images and CLI

### Docker images
When a branch release-candidate or release are created, Build Docker workflow are trigger.

It will create and push docker image, tags with semver, on private docker registry `registry.ogree.ditrit.io`

Docker iamges created are:
- mongo-api/x.x.x: image provide by API/Dockerfile
- ogree-app/x.x.x: image provide by APP/Dockerfile


### CLI

CLI will be build and push into ogree nextcloud folder `/bin/x.x.x/`

### Sermver for docker images and CLI

If build workflow is trigger by a release-candidate branch, workflow will add `.rc` after semver

- release-candidate/1.0.0 will made mongo-api/1.0.0.rc by example

If build workflow is trigger by a release bracnh, workflow will tag OGree-Core with semver

## Secrets needs

- NEXT_CREDENTIALS: nextcloud credentials
- TEAM_DOCKER_URL: Url of the docker registry
- TEAM_DOCKER_PASSWORD: password of the docker registry
- TEAM_DOCKER_USERNAME: username of the docker registry
- GITHUB_TOKEN: an admin github token ( required to trigger build workflow since 08/2022 )