<div align="center">
<img src="https://raw.githubusercontent.com/ditrit/OGrEE-Core/main/APP/assets/custom/logo.png" width="300" alt="NetBox logo" />
<p><strong>A smart datacenter digital twin</strong></p>

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=helderbetiol_OGrEE-Core&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=helderbetiol_OGrEE-Core)
[![⚙️ Build - Publish](https://github.com/ditrit/OGrEE-Core/actions/workflows/build-deploy.yaml/badge.svg)](https://github.com/ditrit/OGrEE-Core/actions/workflows/build-deploy.yaml)

</div>

OGrEE exists so managing a datacenter can be easier. It aggregates data from different datacenter tools to create a digital twin easily accessible from your computer, smartphone or even VR/AR glasses. Its analysis capabilities helps identify issues faster, better prepare for maintenance and migrations, minimize risks and errors.



## What is OGrEE-Core?
OGrEE-Core assembles 3 essential components of OGrEE.

## What is OGrEE-3D?


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

![Workflows diagram](/assets/images/actions.png)

The development process begins with the creation of a new [issue](https://github.com/ditrit/OGrEE-Core/issues). Issues can be created for things such as reporting a bug or requesting a new feature. To work on an issue, a new dedicated branch must be created, and all code changes must be commited to this new branch (**never commit directly to main!**).

Once the development is complete, a [pull request](https://github.com/ditrit/OGrEE-Core/pulls) can be opened. The opening of a pull request automatically triggers two Github workflows: `Branch Naming` and `Commit name check`. If the pull request involves changes to the API, APP and/or CLI, `🕵️‍♂️ API Unit Tests`, `🕵️‍♂️ APP Unit Tests` and/or `🕵️‍♂️ CLI Unit Tests` workflows are also automatically triggered to build and test the changes. If the pull request involves changes to the Wiki, the `📚 Verify conflicts in Wiki` workflow is automatically triggered to check the changes.

If all checks are performed successfully, another member of the team will perform a code review on the pull request. If no further changes are requested, the pull request can be merged into the [main branch](https://github.com/ditrit/OGrEE-Core/tree/main) safely, closing all related issues. The merging of a pull request involving changes to the API, APP and/or CLI will automatically trigger the `🆕 Create Release Candidate` workflow, explained in the [Release candidate](#release-candidate) section below. The merging of a pull request involving changes to the Wiki will automatically trigger the `📚 Publish docs to Wiki` workflow.

## Release candidate

![Release candidate diagram](/assets/images/main.jpg)

After merging a dev branch into main, the `🆕 Create Release Candidate` workflow will create a new branch named `release-candidate/x.x.x`.

Semver bump are defined by the following rules:
- One commit between last tag and main contains: break/breaking -> Bump major version;
- One commit between last tag and main contains: feat/features -> Bump minor version;
- Any other cases -> Bump patch version.

If a branch release-candidate with the same semver already exists, it will be deleted and recreated from the new commit.

Example: A patch is merged after another, which has not yet been released.

This workflow will automatically trigger the `⚙️ Build - Publish` workflow. This workflow is responsible for building the binaries of the API, BACK and CLI (for Windows, MacOS and Linux), the WebAPP, and the Windows Installer, which includes the Windows API, APP, CLI and 3D packet binaries). All of these binaries are then published into [OGrEE's Nextcloud](https://nextcloud.ditrit.io/index.php/apps/files/?dir=/Ogree&fileid=2304). The `⚙️ Build - Publish` workflow is also responsible for building Docker Images for the API, WebAPP and BACK and for publishing these images into OGrEE's private Docker Registry `registry.ogree.ditrit.io`.

## Release

After validating a release candidate, the `📦 Create Release` workflow can be manually run from the [Github Actions panel](https://github.com/ditrit/OGrEE-Core/actions) on the release-candidate branch. This workflow will create a new branch named `release/x.x.x`.

![Github Actions panel](/assets/images/github.png)

Note: If release workflow is launch on another branch other than a release-candidate, it will fail.

Besides creating a new [Github Release](https://github.com/ditrit/OGrEE-Core/releases) for the project, this workflow will also automatically trigger the `⚙️ Build - Publish`, explained in the [Release candidate](#release-candidate) section above. 

## Build docker images and CLI

### Docker images
When a branch release-candidate or release are created, the `⚙️ Build - Publish` workflow will automatically trigger workflows for creatinh the Docker Images, tags with semver, into the private Docker Registry `registry.ogree.ditrit.io`.

Docker images created are:
- mongo-api/x.x.x: image provided by API/Dockerfile;
- ogree-app/x.x.x: image provided by APP/Dockerfile;
- ogree_app_backend/x.x.x: image provided by BACK/app/Dockerfile.

### CLI

CLI will be built and pushed into [OGrEE's Nextcloud](https://nextcloud.ditrit.io/index.php/apps/files/?dir=/Ogree&fileid=2304) folder `/bin/x.x.x/`

### Sermver for Docker Images and CLI

If the build workflow is triggered by a release-candidate branch, the workflow will add `.rc` after semver.

- Example: release-candidate/1.0.0 will be made mongo-api/1.0.0.rc

If the build workflow is triggered by a release branch, the workflow will tag OGrEE-Core with semver.

## Secrets needs

- NEXT_CREDENTIALS: nextcloud credentials
- TEAM_DOCKER_URL: Url of the docker registry
- TEAM_DOCKER_PASSWORD: password of the docker registry
- TEAM_DOCKER_USERNAME: username of the docker registry
- PAT_GITHUB_TOKEN: a personal access github token (required to trigger build workflows)
- GITHUB_TOKEN: an admin github automatic token
