# Gitflow-CLI

[![build](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/build.yml/badge.svg)](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/build.yml)
[![blackduck](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/blackduck.yml/badge.svg)](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/blackduck.yml)

The **gitflow-cli** is a release workflow automation tool built on the [gitflow](https://nvie.com/posts/a-successful-git-branching-model/) branching model. It detects your project type and automatically increments the [semantic version](https://semver.org/), supporting a growing list of [plugins](#available-plugins) out of the box.

<img src=".github/assets/gitflow-cli-demo.png" alt="gitflow-cli-demo" width="600" />

## Installation

From within the project directory the **gitflow-cli** can be built, run and installed.

1. **Clone the repository:**

    ```bash
    git clone https://github.com/mercedes-benz/gitflow-cli.git
    cd gitflow-cli
    ```

2. **Install the application:**

   To install and run the application, use the following commands:

   ```bash
   go install
   gitflow-cli --help
   ```

   Make sure you have [Go](https://go.dev/doc/install) installed and that the `go/bin` directory is part of your PATH.

## Usage

Before using **gitflow-cli**, either navigate to your target Git repository or specify it with the `--path` flag.
Make sure the repository meets all [preconditions](#preconditions).

### Feature

Feature branches are where the work happens. They branch off `develop` and merge back into it,
so a feature that is not finished in time simply stays out of the next release.

To start a new feature, pass its name:

   ```bash
   gitflow-cli feature start 142-book-a-service-appointment
   ```

Feature start will perform the following steps:

* Refuse a feature whose branch already exists, locally or on the remote
* Pull the `develop` branch, so the feature does not start from a stale state. A `develop` that
  exists only locally, for example after an earlier run with `--no-push`, is used as it is
* Create and check out a feature branch from `develop` (e.g., `feature/142-book-a-service-appointment`)
* Push that branch, and only that branch

The name is the part after the prefix. Passing the prefix as well (`feature/142-...`) is accepted and
not doubled, and a name that would shadow a long-lived branch or another workflow's prefix
(`main`, `develop`, `release/...`, `hotfix/...`) is rejected.

Once the feature is ready, finish it with:

   ```bash
   gitflow-cli feature finish
   ```

Feature finish will perform the following steps:
* Pull the feature branch, so commits that others pushed to it are part of the merge
* Pull the `develop` branch
* Merge the feature branch into `develop` with a merge commit, so the individual commits of the
  feature stay visible in the history
* Delete the feature branch locally
* Push `develop`
* Delete the feature branch on the remote, but only while every commit on it is in `develop`.
  If someone pushed to it in the meantime, the branch is kept and the command fails, so the
  new commits can be merged by hand

Both commands pull with a merge and never rebase, whatever `pull.rebase` or `pull.ff` are set to
in the git configuration, so the merge commits of finished features stay in `develop`.

Without an argument the feature branch that is currently checked out is finished. Naming it
explicitly (`gitflow-cli feature finish 142-book-a-service-appointment`) picks a specific one.

A feature branch never carries a version: the version is only moved by the release and hotfix
workflows. Nothing here touches the version file, so the feature commands need no build tool and no
Docker, only `git`.

Note that `feature finish` merges locally. Where changes reach `develop` through a reviewed merge
request, use the merge request instead of this command.

### Release

To initiate a new release, use the following command:

   ```bash
   gitflow-cli release start
   ```

Release start will perform the following steps:

* Create a new release branch from `develop` (e.g., `release/1.2.0`)
* Remove the version qualifier in the version file (e.g., `1.2.0-dev` → `1.2.0`)

You can now use the `release/x.y.z` branch for bug fixing, creating the release changelog, or deploying your app to your testing environment.

Once the release is ready, finish it with:

   ```bash
   gitflow-cli release finish
   ```

Release finish will perform the following steps:
* Merge the `release/x.y.z` branch into `main` (e.g., `release/1.2.0` → `main`)
* Create a tag in `main` with the corresponding version (e.g., `1.2.0`)
* Perform a back-merge into `develop` (e.g., `release/1.2.0` → `develop`)
* Bump the development version to the next minor version (e.g., `1.3.0-dev`)

### Hotfix

Use hotfixes if you have a bug in production, and you need to make targeted fixes to `main` branch without deploying pending changes from `develop`.

To initiate a new hotfix, use the following command:

   ```bash
   gitflow-cli hotfix start
   ```

Hotfix start will perform the following steps:
* Create a `hotfix/x.y.z` branch from `main` (e.g., `hotfix/1.2.1`)
* Set the patch version in the version file (e.g., `1.2.0` → `1.2.1`)

You can now check out the `hotfix/x.y.z` branch, create a quick patch, and push your changes.

Once the hotfix is ready, finish it with:

   ```bash
   gitflow-cli hotfix finish
   ```

Hotfix finish will perform the following steps:
* Merge the `hotfix/x.y.z` branch into `main` (e.g., `hotfix/1.2.1` → `main`)
* Create a tag in `main` with the corresponding version (e.g., `1.2.1`)
* Perform a back-merge into `develop` (e.g., `hotfix/1.2.1` → `develop`)
* Keep the current version in `develop` unchanged (e.g., `1.3.0-dev`)

## Preconditions

To use **gitflow-cli**, ensure your project meets the basic structural requirements, particularly around Git branches and version management.

### Prerequisites

- **git** — required for all operations, and the only requirement for the feature workflow

- **Native Mode** (`--native-mode`, default)
  - The respective build tool (e.g., `mvn`, `npm`, `composer`, `toml`) must be installed and available in PATH.
  - If the native tool is missing, Docker is used automatically as fallback.

- **Docker Mode** (`--docker-mode`)
  - Only [Docker](https://docs.docker.com/get-docker/) needs to be installed — no build tools required on the host.
  - Commands run inside disposable containers (removed automatically after each invocation).

### Git Branches

Your repository must define a dedicated **production** and **development** branches (e.g., `main` and `develop`).
These can be [customized](#configuration) as needed.

### Version File

Each project type may store version information in a different location.
The **gitflow-cli** detects your project's context and automatically delegates tasks to the appropriate plugin based on the presence of specific file.

#### Available Plugins

| Plugin       | Description                                                                                      | Required File                                 |
|--------------|--------------------------------------------------------------------------------------------------|-----------------------------------------------|
| **standard** | Plugin for projects without a dedicated version file.                                            | `version.txt`                                 |
| **mvn**      | Plugin for [maven](https://maven.apache.org) projects.                                           | `pom.xml`                                     |
| **npm**      | Plugin for [npm](https://www.npmjs.com/) projects.                                               | `package.json`                                |
| **python**   | Plugin for [python](https://www.python.org/) projects.                                           | `pyproject.toml` \| `setup.cfg` \| `setup.py`    |
| **composer** | Plugin for [composer](https://getcomposer.org/) projects.                                        | `composer.json`                               |
| **road**     | Plugin for projects with road app manifest configuration.                                        | `road.yaml`                                   |


If no technology-specific plugin can be applied, **gitflow-cli** will create a `version.txt` file in your project's root directory and apply the **standard** plugin.

## Configuration

A configuration file is automatically created at `$HOME/.gitflow-cli.yaml` on first run. You can also specify a custom path with `--config`.

### Configuration Reference

```yaml
branches:
  production: main       # Name of the production branch
  development: develop   # Name of the development branch
  release: release       # Prefix for release branches
  hotfix: hotfix         # Prefix for hotfix branches
  feature: feature       # Prefix for feature branches

workflow:
  push: true             # Push changes to remote after workflow completes
  rollback: false        # Rollback local changes on workflow failure
  docker-fallback: true  # Automatically use Docker when native tool is missing

logging: "off"           # Diagnostic output (combinable: stdout, stderr, cmdline, output, off)
```

Values are resolved in order: CLI flag → config file → default.


## Contributing

We welcome any contributions.
If you want to contribute to this project, please read the [contributing guide](CONTRIBUTING.md).

### Git Hook

To contribute to **gitflow-cli**, we suggest setting up the Git hook below to comply with our contribution guidelines.

   ```bash
   cp .githooks/prepare-commit-msg .git/hooks/
   chmod +x .git/hooks/prepare-commit-msg
   ```

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) as it is our base for interaction.

## License

This project is licensed under the [MIT LICENSE](LICENSE).

## Provider Information

Please visit <https://www.mercedes-benz-techinnovation.com/en/imprint/> for information on the provider.

Notice: Before you use the program in productive use, please take all necessary precautions,
e.g. testing and verifying the program with regard to your specific use.
The source code has been tested solely for our own use cases, which might differ from yours.
