# Tutorial: Creating a Course

TODO
In this tutorial ...
Start from no knowledge of the autograder.
Will run using Docker.
Will create a course and assignment from scratch.

Requirements:
 - Edit files.
 - Docker (for both running the server and graders).
   - For non-Docker usage, see the:
     - Readme for running the server without Docker (TODO)
     - Readme for running graders without Docker (TODO)
 - Python (for scripts)
 - POSIX System (e.g., Linux, BSD, Mac, WSL (Windows))

## Basic Server Control

In this section, we will cover the basics of starting up and controlling an autograder server.
Although you may already have a server deployed for you and your class,
being able to use a server is critical for creating and testing your course.

Although there are several ways to run an autograder server (build from source, use prebuilt binary, etc),
Although there are several ways to run autograder commands (build from source, use prebuilt binary, etc),
this guide will only cover using the existing Docker image.
Using the prebuilt Docker image is the easiest and most consistent way to run autograder commands (including the server).
If you need instructions on installing or using Docker, see the [Official Docker documentation](https://docs.docker.com/desktop/).
There are also many additional resources online to help you get started with docker.

In this guide, we will make heavy use of a script to run the Docker container for us.
This script sets various options automatically.
To learn more about this script, see this [project's Docker documentation](../docker.md) or use the `--help` flag:
```sh
./docker/run-docker-server.py --help
```

This script (`./docker/run-docker-server.py`) is useful for running any command in [the cmd directory](../../cmd) inside a docker container.
For example, you can check the version of your server using:
```sh
./docker/run-docker-server.py version
```

### Starting a Test Server

TODO: Do we even need this?

All the details of starting the server with Docker are detailed in this project's [Docker documentation](../docker.md#running-the-server),
but the easiest way is to use our Docker Python script:
```sh
./docker/run-docker-server.py server
```

You should see some logging output (similar to the following) indicating that the server is running:
```
2024-10-20T21:02:16.692Z [ INFO] Autograder Version. | {"version":{"short-version":"0.3.0","git-hash":"c6f76d8e","is-dirty":false,"api-version":3}}
2024-10-20T21:02:16.695Z [ INFO] API Server Started. | {"port":8080}
2024-10-20T21:02:16.696Z [ INFO] Unix Socket Server Started. | {"unix_socket":"/tmp/autograder-4695a16cff52cec7b599b2c1e1e4f2c3.sock"}
```

When testing/debugging your course, you may also find it useful to set your logging level to debug:
```sh
./docker/run-docker-server.py server -- --log-level DEBUG
```
Note the `--` tells the script that everything after it is an argument to the autograder server and not the script itself.

To stop the server, just use Ctrl-C.

To run the server with preloaded courses and users (used for testing), use the `--unit-testing` flag:
```sh
./docker/run-docker-server.py server --unit-testing
```

### Querying Logs

TODO

### Clearing the Database

TODO

**WARNING**: This command cannot be undone, only do this in an environment you own.

## Configuration

TODO: We may not need this (if we are not using the server).

In this section, we will cover some of the configuration you may want to set for your test server.
For a more in-depth discussion of configuration for the autograder, see [this document](../../README.md#configuration).

TODO: setting config values

```sh
./docker/run-docker-server.py list-options
```

TODO: base dir

## Making a Course

Now that we can run autograder commands, let's make a course!
In this guide, we will make a course called "my-first-course".
A complete instantiation of "my-first-course" is available in the [docs/tutorials/resources/my-first-course](my-first-course) directory.
We encourage you to make each file on your own as you progress through this guide,
but the full implementation is available for reference.

### Course Configuration

A course is defined by a `course.json` file, referred to as the "course configuration".
The directory containing the `course.json` file is referred to as the "course directory",
and is the root/base directory for a course.
At its simplest, a course configuration is just required to define the ID of the course.
See the [course configuration documentation](../types.md#course) for a list of all the possible configuration options.

Let's make a course called `my-first-course`.
To do so, we will first need to make a directory for it,
and then create a `course.json` file for it that contains an "id" field:
```console
edulinq@example:autograder-server$ mkdir my-first-course
edulinq@example:autograder-server$ cd my-first-course
edulinq@example:autograder-server$ echo -e '{\n    "id": "my-first-course"\n}' > course.json
edulinq@example:autograder-server$ cat course.json
{
    "id": "my-first-course"
}
```

TODO: Adding a course to the server

TODO: Source / Updating

TODO: Tasks

## Making an Assignment

TODO: Config

TODO: Assignment Dir

TODO: Graders - Basic Concept

The "grader" is the program responsible for looking at a student's submission and assigning a score to it.
Naturally, this is typically the most complex and specialized component of creating a course for the autograder.
When a grader finishes running, it is supposed to produce a JSON file representing a [grading result object](TODO) to `/output/result.json`.
Graders may be created using any language or library you can run in your Docker image.
If you have existing grading scripts/programs,
it should be fairly simple to modify or wrap them so that it creates the required JSON file when it finishes.

In this tutorial, we will create a simple grader using the canonical [Python Interface Library](https://github.com/edulinq/autograder-py) for the autograder.
However, you can make graders in any language (that can run in Docker) and use whatever libraries you want to help you grade.
One way to think of graders is like unit tests with additional feedback and partial credit.
Here are some examples of other autograder graders:
 - [Regex Grader](https://github.com/edulinq/cse-cracks-course/blob/main/assignments/regex/grader.py)
   - This is a fully featured and commented grader used for an open source assignment on Regular Expressions.
    - It uses the canonical [Python Interface Library](https://github.com/edulinq/autograder-py) for the autograder.
      - This provides a lot of nice functionality like timeouts, automatic importing, handling of Jupyter Notebooks, etc.
 - These graders are used in the unit tests for this repo (so they are fairly simple).
  - [Python Grader using autograder-py Library](../../testdata/course101/HW0/grader.py)
  - [Bash Grader](../../testdata/course-languages/bash/grader.sh)
  - [C++ Grader](../../testdata/course-languages/cpp/grader.cpp)
  - [Java Grader](../../testdata/course-languages/java/Grader.java)

TODO: Base Image

Solution
Serves as a good test of what correct code should look like.

Not implemented
Ensure that a student submitting a blank submission gets the code you expect.

Syntax Error
Great for testing your own grading infrastructure.
Students will inevitably submit bad code, ideally your grader should be able to handle it without crashing.

python3 -m autograder.cli.testing.test-submissions --submission test-submissions

TODO: Build / Run Stages

TODO: Files

TODO: Testing Graders
 - It's a good idea to always make graders easy to run locally.
 - Test Submissions
 - Include Running Graders in CI

TODO
Highlight important things I'm grader
Fallback Jason
Run alone
Test submissions (link to sample course ci)

TODO
Recommended test submissions
Solution
Empty
No compile / bad syntax
