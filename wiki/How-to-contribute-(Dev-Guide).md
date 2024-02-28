Hi there! So you want to contribute to this projects? That's great! Let's go over some basic points to help you get started.

## Workflow
This is a **mono repo** (aka, a lot of things in the same place) for the Core applications of OGrEE: API, CLI and BACK, all developed in Go, as well as APP, developed in Flutter. To organise all this, we follow a **feature branching** inspired workflow:

* First, create an **issue**: this can report a bug, ask for a new functionality, propose discussions or ask questions. 

* If development is needed, create a **branch**. The name of the brach should be something like `{keywork}{#issue}-quick-description`, where it starts with one of the following keywords: `feature`, `feat`, `fix`, `hotfix`, `release`, `chore`, `break`, `breaking`, `docs`, followed by the number of the issue and a few words to describe its purpose. For example, if you created na issue to report a bug in the APP's login page, the branch name could be: `fix123-app-login-issue`.

* Many new features of even bug fixes may impact more than one application at a time. All the work is done in the same branch. So if your issue impacts the CLI and API at the same time, for example, both should be changed in the same branch. Each **commit** should also follow a naming pattern: `{keyword}({application}) quick description`, where the keyword is one of the following: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`, `break`, `breaking`, `hotfix`. Example: `fix(app) handle login exception`.

> Be aware: our GitHub workflows will check the branching and commit naming patterns when a PR is created ;)

* Once all work is done, create a **pull request** (PR). This will trigger several Github workflows (more about it [here](https://github.com/ditrit/OGrEE-Core?tab=readme-ov-file#-ogree-core-gitflow)) that will test and build your code. If all it's good, it's time for **peer review**! Another member of the time will check your code, maybe ask some questions or propose some changes. Once approved, the only thing left is to squash and merge. Your code will get to the main branch, closing the issue and triggering more [Github workflows](https://github.com/ditrit/OGrEE-Core?tab=readme-ov-file#-ogree-core-gitflow) to create or aggregate to a release candidate version. 

 
 