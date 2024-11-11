# Tutorial: Creating a Course

TODO
In this tutorial ...
Start from no knowledge of the autograder.
Will run using Docker.
Will create a course and assignment from scratch.

Requirements:
 - Edit files.
   - Most config files are [JSON](https://en.wikipedia.org/wiki/JSON).
 - Docker (for both running the server and graders).
   - For non-Docker usage, see the:
     - Readme for running the server without Docker (TODO)
     - Readme for running graders without Docker (TODO)
 - Python (for scripts)
 - POSIX System (e.g., Linux, BSD, Mac, WSL (Windows))

## Running Autograder Commands

In this tutorial, we will be interacting with the autograder using a prebuilt Docker image to access admin commands for the autograder.
Specifically, we will be using a Python script to run a Docker container to run autograder commands
(Python Script -> Docker Container -> Autograder Commands).
This may seem a bit confusion, but it allows you (the reader) to get started with the autograder without needing to compile anything.
We do require both Python (>= 3.8) and Docker to be installed.

The Docker script we are using will set various options automatically.
To learn more about this script, see this [project's Docker documentation](../docker.md) or use the `--help` flag:
```sh
./docker/run-docker-server.py --help
```

Note that all commands should be run from this project/repository's root directory (the one with [README.md](../../README.md) in it).

This script (`./docker/run-docker-server.py`) is useful for running any command in [the cmd directory](../../cmd) inside a docker container.
We refer to any executable in this directory a "CMD", and they are usually used as admin autograder commands.
The autograder (including these CMDs) are written in the [Go programming language](https://en.wikipedia.org/wiki/Go_(programming_language)).
For example, you can check the version of your server using the [version CMD](../../cmd/version/main.go):
```sh
./docker/run-docker-server.py version
```

Although we will be using the CMDs to interact with the autograder in this tutorial,
most regular users will interact with the autograder via the [Python API](https://github.com/eriq-augustine/autograder-py)
or directly via the [Rest API](../../resources/api.json).
If you choose, you can complete this entire tutorial via those interfaces as well.

## Making a Course

Now that we can run autograder commands, let's make a course!
In this guide, we will make a course called "my-first-course".
A complete instantiation of "my-first-course" is available in the [docs/tutorials/resources/my-first-course/final](my-first-course/final) directory.
As we go along, we will make various other versions of this course to demonstrate other things.
We encourage you to make each/edit file on your own as you progress through this guide,
but full implementations are available for reference.

A course is specified by a directory containing a JSON file called `course.json`.
We call the `course.json` file the "course config",
and the directory that the files lives in the "course directory".

To start our new course let's make an empty directory and create a `course.json` file in it.
To start, a course config just needs to specify an identifier for your course:
```json
{
    "id": "my-first-course"
}
```

Once we have our `course.json` file, we can add it to our local autograder using the `upsert-course-from-filespec` CMD:
```sh
./docker/run-docker-server.py \
    --mount docs:/tmp/docs \
    upsert-course-from-filespec \
    /tmp/docs/tutorials/resources/my-first-course/initial-course
```

You will get some output like:
```JSON
[
    {
        "course-id": "my-first-course",
        "success": true,
        "message": "",
        "created": true,
        "updated": false,
        "lms-sync-result": null,
        "built-assignment-images": []
    }
]
```

Let's break apart this command a bit to understand it:
 - `./docker/run-docker-server.py` -- The script we are running. We have seen this before.
 - `--mount docs:/tmp/docs` -- This tells the script that we need to mount our docs directory inside our Docker container. This allows us to accsss data on our host machine (like our new course) inside our Docker container.
 - `upsert-course-from-filespec` -- The CMD we are running.
 - `/tmp/docs/tutorials/resources/my-first-course/initial-course` -- The path (inside our Docker container) to our course we are adding.

Note that we will be using various differnt directories that hold versions of this tutorial course as we build it up from scratch.
You can use these as you go through the tutorial,
or you can build your course yourself as you progress through this tutorial and use the same directory each time.





For a complete description of all the fields a course can have,
see the [course section of the types documentation](../types.md#course).

TEST


Note that all JSON config files in the autograder are strict JSON and don't accept things like trailing commas or unquoted fields.
Let's see what it looks like when we use a config with bad JSON:
```sh
./docker/run-docker-server.py \
    --mount docs:/tmp/docs \
    upsert-course-from-filespec \
    /tmp/docs/tutorials/resources/my-first-course/bad-json
```

We get some output that contains the string: "Could not unmarshal JSON file".







This command can be used both to create and update courses, so you can run it over and over.
Notice that if you run it again, the output will tell you the course was updated rather than created.
If you do want to reset your database, you can use the `clear-db` command:
```sh
./docker/run-docker-server.py clear-db
```

Outputs:
```JSON
[
    {
        "course-id": "my-first-course",
        "success": true,
        "message": "",
        "created": false,
        "updated": true,
        "lms-sync-result": null,
        "built-assignment-images": []
    }
]
```

## Making an Assignment

Now that we have a course, let's make an assignment!
Like a course, as assignment is specified by a JSON config file: `assignment.json`.
Any directory within a course directy that has as `assignment.json` file is considered a base directory for that assignment.
The hard requirements for an assignment are the assignment config and grader.

As assignment's "grader" is the program responsible for looking at a student's submission and assigning a score to it.
Naturally, this is typically the most complex and specialized component of creating a course for the autograder.
When a grader finishes running, it is supposed to produce a JSON file representing a [grading output object](../types.md#grader-output) to `/output/result.json`.
Graders may be created using any language or library you can run in your Docker image.
One way to think of graders is like unit tests with additional feedback and partial credit.
If you have existing grading scripts/programs,
it should be fairly simple to modify or wrap them so that it creates the required JSON file when it finishes.
In this tutorial, we will create a simple grader using the canonical [Python Interface Library](https://github.com/edulinq/autograder-py) for the autograder.

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

First, we need to create an assignment config:
```json
{
    "id": "assignment-01",
    "static-files": [
        "grader.py"
    ],
    "image": "edulinq/grader.python-alpine"
}
```

Let's look at each field:
 - `id` -- The identifier for the assignment, similar to a course identifier.
 - `static-files` -- This is a list of files the grader needs. This can simply be relative paths (starting at the assignment directory), or [more complex file specifications](../types.md#file-specification-filespec) that can point to online resources.
 - `image` -- The base Docker image to use for our grader.

To see the full specification of the fields for an assignment config,
see the [assignment section of the types documentation](../types.md#assignment).

Now let's create the grader, which (as you can see in `static-files`) should be named `grader.py`.
For breviety we won't put the full source code in this document,
but you can see is at
[resources/my-first-course/initial-assignment/assignment-01/grader.py](resources/my-first-course/initial-assignment/assignment-01/grader.py).

Now that we have a full assignment, let's add our course again and ensure that there are no errors:
```sh
./docker/run-docker-server.py \
    --mount docs:/tmp/docs \
    upsert-course-from-filespec \
    /tmp/docs/tutorials/resources/my-first-course/initial-assignment
```

In our output we can see that the autograder updated our course,
and successfully built Docker images for our new assignment!
```json
[
    {
        "course-id": "my-first-course",
        "success": true,
        "message": "",
        "created": false,
        "updated": true,
        "lms-sync-result": null,
        "built-assignment-images": [
            "autograder.my-first-course.assignment-01"
        ]
    }
]
```









TEST




Note that our assignment config is missing the `invocation` field that tells the autograder how to run the grader program/script.
For canonical Python grader images (`edulinq/grader.python-*`),
the autograder already understands how to unvoke those.
If you put in your invocation manually, it would be:
```json
{
    ... the rest of an assignment object ...
    "invocation": "python3 -m autograder.cli.grading.grade-dir --grader grader.py --dir /autograder/work --outpath /autograder/output/results.json"
}
```


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

---
TEST






### Querying Logs

The autograder logs events that happen on the server.
For a course builder, it is important that you are able to recover any warnings or errors that are logged about your course.
To do that, you can query server logs using the `logs-query` command or `autograder.cli.logs.query` Python interface.

logs-query command
```sh
./docker/run-docker-server.py logs-query -- --unit-testing --log-level DEBUG
```

or

Python
```sh
python -m autograder.cli.logs.query --target-course course101 --user server-admin@test.edulinq.org --pass server-admin --server http://127.0.0.1:8080 --level DEBUG
```


mostly relevant 

[logging](../types.md#logging)

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






---





## Basic Server Control

In this section, we will cover the basics of starting up and controlling an autograder server.
Although you may already have a server deployed for you and your class,
being able to use a server is critical for creating and testing your course.

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

### Interacting with the Autograder



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

### Connecting to Your Server

TODO

Python API

