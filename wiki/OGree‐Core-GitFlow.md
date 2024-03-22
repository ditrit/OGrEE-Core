
![Workflows diagram](https://github.com/ditrit/OGrEE-Core/blob/main/assets/images/actions.png)

The development process begins with the creation of a new [issue](https://github.com/ditrit/OGrEE-Core/issues). Issues can be created for things such as reporting a bug or requesting a new feature. To work on an issue, a new dedicated branch must be created, and all code changes must be commited to this new branch (**never commit directly to main!**).

Once the development is complete, a [pull request](https://github.com/ditrit/OGrEE-Core/pulls) can be opened. The opening of a pull request automatically triggers two Github workflows: `Branch Naming` and `Commit name check`. If the pull request involves changes to the API, APP and/or CLI, `🕵️‍♂️ API Unit Tests`, `🕵️‍♂️ APP Unit Tests` and/or `🕵️‍♂️ CLI Unit Tests` workflows are also automatically triggered to build and test the changes. If the pull request involves changes to the Wiki, the `📚 Verify conflicts in Wiki` workflow is automatically triggered to check the changes.

If all checks are performed successfully, another member of the team will perform a code review on the pull request. If no further changes are requested, the pull request can be merged into the [main branch](https://github.com/ditrit/OGrEE-Core/tree/main) safely, closing all related issues. The merging of a pull request involving changes to the API, APP and/or CLI will automatically trigger the `🆕 Create Release Candidate` workflow, explained in the [Release candidate](#release-candidate) section below. The merging of a pull request involving changes to the Wiki will automatically trigger the `📚 Publish docs to Wiki` workflow.

## Release candidate

![Release candidate diagram](https://github.com/ditrit/OGrEE-Core/blob/main/assets/images/main.jpg)

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

![Github Actions panel](https://github.com/ditrit/OGrEE-Core/blob/main/assets/images/github.png)

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