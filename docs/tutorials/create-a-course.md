# Tutorial: Creating a Course

TEST -- my-autograder-server-prebuilt

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



TODO



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
When a grader finishes running, it is supposed to produce a JSON file representing a [grading output object](../types.md#grader-output-graderoutput) to `/output/result.json`.
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
For brevity we won't put the full source code in this document,
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

*Sidenote:*  
Note that our assignment config is missing the `invocation` field that tells the autograder how to run the grader program/script.
For canonical Python grader images (`edulinq/grader.python-*`),
the autograder already understands how to invoke those.
If you put in your invocation manually, it would be:
```json
{
    ... the rest of an assignment object ...
    "invocation": "python3 -m autograder.cli.grading.grade-dir --grader grader.py --dir /autograder/work --outpath /autograder/output/results.json"
}
```

### Testing an Assignment

We no have an assignment with a grader that successfully creates a Docker image,
but we have not actually tested that the grader does what we want it to do.
Let's make a few sample submissions and check the output.

*Sidenote:*
Note that in a full course, we would create [test submissions](#test-submission),
which are more formal ways to define the expected behavior in a grader.
But, that is outside the scope for this tutorial.
Here are some examples of test submissions:
 - [autograder-server test course](https://github.com/edulinq/autograder-server/tree/main/testdata/course101/HW0/test-submissions).
   - These test submissions are part of one of the test courses used for unit testing in this project.
 - [Regex Assignment](https://github.com/edulinq/cse-cracks-course/tree/main/assignments/regex/test-submissions)
   - These test submissions are a part of a fully featured sample assignment/course.
   - This repository highlights good practices with autograder courses, including running these test submissions are part of [continuous integration](https://en.wikipedia.org/wiki/Continuous_integration)
     - [CI Script that Runs the Test Submissions](https://github.com/edulinq/cse-cracks-course/blob/main/.ci/check_graders.sh)

We will create three sample submissions for this tutorial in the [resources/my-first-course/final/assignment-01/sample-submissions](resources/my-first-course/final/assignment-01/sample-submissions) directory:
 - [not-implemented](resources/my-first-course/final/assignment-01/sample-submissions/not-implemented/submission.py)
   - This submission represents the initial code that a student is given to start the assignment.
     Including a sample submission like this makes it easy to check that a student the number of expected points for submitting a blank assignment.
     (We don't want to accidentally give a student more points for an empty solution than a wrong solution.)
 - [solution](resources/my-first-course/final/assignment-01/sample-submissions/solution/submission.py)
   - This submission represents the ideal submission a student could give.
     Including a sample submission like this makes lets you ensure that a perfect solution gets a full score.
     If there are multiple possible solutions, including multiple sample solutions is a good idea.
 - [syntax-error](resources/my-first-course/final/assignment-01/sample-submissions/syntax-error/submission.py)
   - This submission represents s submission that does not compile/parse/run.
     Dealing with student submissions that crashes is a hard part of writing a grader.
     The autograder will catch these situations and give the student a zero for that submission,
     but catching these situations in the grader itself will allow you to give more detailed feedback.

Now that we have sample submissions, let's run them through the autograder using the `grade` CMD.
First let's start with the `not-implemented` submission:
```sh
./docker/run-docker-server.py \
    --image my-autograder-server-prebuilt \
    --mount docs:/tmp/docs \
    grade \
    -- \
    my-first-course assignment-01 \
    --submission /tmp/docs/tutorials/resources/my-first-course/final/assignment-01/sample-submissions/not-implemented
```

You should see output like:
```
Autograder transcript for assignment: Assignment1.
Grading started at 2024-01-01T01:01:01.001Z and ended at 2024-01-01T01:01:01.001Z.
Question 1: Add: 1 / 10
    Add() did not return a correct result on test case 'zeros'.
    Add() did not return a correct result on test case 'basic'.
    Add() did not return a correct result on test case 'negative'.

Total: 1 / 10
```

As expected, this submission does not get a very good score.
If you look closely at our grader you can see the submission gets a single point via integer division.

Now let's try the `solution` submission:
```sh
./docker/run-docker-server.py \
    --image my-autograder-server-prebuilt \
    --mount docs:/tmp/docs \
    grade \
    -- \
    my-first-course assignment-01 \
    --submission /tmp/docs/tutorials/resources/my-first-course/final/assignment-01/sample-submissions/solution
```

Here we can see a clean 100% score:
```
Autograder transcript for assignment: Assignment1.
Grading started at 2024-01-01T01:01:01.001Z and ended at 2024-01-01T01:01:01.001Z.
Question 1: Add: 10 / 10

Total: 10 / 10
```

Finally, let's try out the `syntax-error` submission:
```sh
./docker/run-docker-server.py \
    --image my-autograder-server-prebuilt \
    --mount docs:/tmp/docs \
    grade \
    -- \
    my-first-course assignment-01 \
    --submission /tmp/docs/tutorials/resources/my-first-course/final/assignment-01/sample-submissions/syntax-error
```

Note that the submission could not be graded, but the grader still returns an appropriate result:
```
Autograder transcript for assignment: Assignment1.
Grading started at 2024-01-01T01:01:01.001Z and ended at 2024-01-01T01:01:01.001Z.
Question 1: Add: 0 / 10
    Submission could not be graded.

Total: 0 / 10
```

The autograder's Python library (`autograder-py`) handles this case well,
but this is probably one of the most difficult cases for a custom grader to handle.
The student's code behaving poorly can be very difficult to catch.
Specifically, the following cases can be difficult to manage:
 - Does not compile/parse/run.
 - Takes too long to run.
 - Outputs too much on stdout/stderr.
 - Writes too much to disk.

The autograder deals with all these situations,
but does not have the low-level details to handle them as gracefully as a grader can.
For example, imagine the case of a student's code crashing on just one question.
The autograder will see that the grader has completed but has not written out the grading result file,
so will consider the entire grading a failure and give a zero score.
But, a grader could possibly catch the crash/failure/error/exception for that specific question and continue grading while just assigning a zero for that question.
You can see that this is what the autograder's Python library does.


## Next Steps

TODO

Congrats!
