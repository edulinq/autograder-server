# Autograder Server

[![Build Status](https://github.com/edulinq/autograder-server/actions/workflows/main.yml/badge.svg)](https://github.com/edulinq/autograder-server/actions/workflows/main.yml)

A server for automatically grading programming assignments in real-time.

The autograding effort is broken into three main parts:
 1. The autograding server that can accept student submissions, run the assignment graders, and provide feedback to students.
    This repository implements that server.
 2. The interfaces through which users can interact with the autograding server.
    Since the autograding server interacts via a REST API, the requirements for an interface are very low.
    Currently, the [autograder-py Python package](https://github.com/edulinq/autograder-py) is the only official interface package.
 3. Course and assignment configurations, including graders to score student submissions.
    These materials should generally be kept private (since they include grading information),
    but we have made a [sample course](https://github.com/edulinq/cse-cracks-course) available.

## Quick Links

 - [Running the Server](#running-the-server)
   - [Running the Server in a Docker Container](docs/docker.md#running-the-server)
 - [Configuration](#configuration)
 - [Description of all Autograder Types](docs/types.md)
 - [Tutorial -- Creating a Course](docs/tutorials/create-a-course.md)
 - Resources
   - [Server (this repo)](https://github.com/edulinq/autograder-server)
   - [Docker Images](https://hub.docker.com/u/edulinq)
   - [Python Interface](https://github.com/edulinq/autograder-py)
   - [Sample Course](https://github.com/edulinq/cse-cracks-course)

## Building

This project uses Go 1.21.
Development and deployment of this project rely on POSIX systems (e.g., Linux, macOS, WSL).

The built-in web GUI is stored in a git submodule,
so you should clone with `--recurse-submodules` if you need that component.
See the [Git documentation](https://git-scm.com/book/en/v2/Git-Tools-Submodules) for more information.

All code that is not intended to be exported (used in packages outside of the autograder) is in the `internal` package/directory.
Since this is a server and not a library, that is the majority of the code.

By default, assignments are graded using Docker.
Therefore when grading functionality is used,
Docker should be installed on the machine and accessible to the current user without additional permissions.
Users without Docker can run the server without Docker (see below).

The project adheres to standard Go standards,
so the `go` tool can be used to build, test, manage, etc.
Additionally, the `scripts/build.sh` script is provided which will build all executables in this project into the `bin` directory.
```sh
./scripts/build.sh
```

All executable mains are kept in the `cmd` directory.
Each includes a usage and responds to the `--help` flag.

Tests can be run using the standard `go test` tool chain.
You can also use the `scripts/run_tests.sh` script for running all tests:
```sh
./scripts/run_tests.sh
```

## Running Executables

Once built, all executables are available in the `bin` directory and can be run directly.
For development, these executables can also be run via `go run`, which will rebuild them if necessary before running:
```
go run cmd/version/main.go
```

## Configuration

The autograder server uses a tiered configuration system that includes
files, environmental variables, and command-line options.
This section will just highlight some key uses.
For more information about autograder configuration (including all config options),
see the [configuration documentation](docs/config.md).

All autograder executables sets their configuration the same way,
so you can use the same files or command-line options for each executable.
In-general, you can set options using the `-c`/`--config` flag.

For example, you can set debug logging with:
```
./bin/logs-example --config log.level=debug

# Or, the short version.
./bin/logs-example --log-level debug
```

## Preparing for Grading

Before the server is ready to grade student submissions,
you may have to take some steps depending on your server and assignments.

### Docker Grading

If using the standard Docker-based grading,
then you should take two steps before starting the server for grading.

First, ensure that any required Docker images are accessible for building.
The required images depend on what courses you are hosting,
but the default images live in the [autograder-docker](https://github.com/edulinq/autograder-docker) repository.

Second, you will want to pre-build your grader images.
This step is not required as the autograder will ensure that grader images are up-to-date before running a grader,
but this will ensure that the first student to submit is not stuck with a long wait.
Building images can be done using the `cmd/build-images` executable,
which will build images for all known assignments.

```
./bin/build-images
```

### Non-Docker Grading

When Docker is not available,
users can choose to run the server without Docker.
Non-Docker grading will only work when the default Python grader is used.
Also note that running the grader without Docker is a potential security risk
and should be avoided in production.

To disable docker, set the `docker.disable` config option to `true`.

The [Python autograder interface](https://github.com/edulinq/autograder-py) must be installed for the non-Docker grader to work:
```
pip install autograder-py
```

## Running the Server

The main server is available via the `cmd/server` executable.
```
./bin/server
```

The `web.port` config option can be used to set the port the server listens on:
```
./bin/server -c web.port=80
```

If you want to run on a privileged port as a non-root user,
we recommend using [setcap](https://man.archlinux.org/man/setcap.8).
The `scripts/setcap.sh` script will do this for you:
```
./scripts/setcap.sh
```

### Running the Server for Testing

You may want to run the server for testing/debugging,
e.g., if you are developing an interface to the server.
We recommend two additional changes to how you would normally run the server:
```
go run cmd/server/main.go --unit-testing
```

First, we ran the server using `go run`,
This will ensure that the server executable is up-to-date before running it.
Second we used the `--unit-testing` flag,
which will set some testing options, create a clean new database, and load the test courses (inside the `testdata directory).

Additionally, when running the server in `--unit-testing` mode,
most configs may get overwritten by the testing infrastructure but environmental variables will not get overwritten.
For more information about config options see the [Configuration section](#configuration) of this document.

## Running Tests

This repository comes with several types of tests.
All these tests are run in the CI,
and can also be run using the `./.ci/run_all.sh` script:
```
./.ci/run_all.sh
```

Users may also choose to run them individually.

### Base Tests

The base tests are created with Go's `testing` standard library package,
and can therefore be run using `go test`.
The `scripts/run_tests.sh` script runs `go test` for each package:
```
./scripts/run_tests.sh
```
