# Tutorial: Troubleshooting a Course

In this tutorial we are going to cover some of the common pitfalls when creating a course and assignment for the autograder.
We hope to help provide the tools necessary to troubleshoot a course and assignment.
This tutorial assumes you have already set up a course and assignment for the autograder.
If not, start by following the [create a course tutorial](create-a-course.md).

## Preface

The easiest way to debug courses and assignments is to set up a foundation of continuous integration.
If your course needs continuous integration, follow the [create continuous integration tutorial](create-continuous-integration.md).

## Simple Fixes

These fixes are quick sanity checks that catch simple issues.
 - Ensure to install all necessary requirements to run the autograder.
 - Upgrade your version of `autograder-py` by running:
```
pip3 install -U autograder-py
```

## Slow Graders

As grading is the core user experience, fast graders are critical, but can be challenging.

Every autograder-server instance has a maximum amount of time that a grading container can run before it is killed.
The amount of time is determined by the [grading.runtime.max](../config.md#configuration-options) (system timeout) configuration on the autograder-server.

Additionally, each assignment can set the [max-runtime-secs](../types.md#assignment) configuration option.
The assignment max runtime cannot be greater than the system limit.

Ensure these values are set as expected to avoid prematurely killing a grading container.
Once you determine the number of seconds a grading container can run, check the following possibilities.

### Infinite Loops / Recursion

Submissions have the possibility to contain infinite loops or recursion.
Graders should be prepared to detect when a question is taking too long to run, which can be quite difficult.

We recommend setting a timeout for each question to kill the thread (or subprocess) that is grading the question.
This allows graders to gracefully report slow (or infinite) implementations to the user.
This way, the infinite loop only ruins a single question not the entire assignment.

Be sure to test your graders properly timeout slow (or infinite) submissions.

### Speeding Up Graders

If ideal solutions are slow or cannot pass the tests within the grading container timeout,
you will need to speed up grading scripts.

While we cannot suggest language or grader based optimizations,
we recommend the grading image complete as much work ahead of receiving a submission as possible.
While this is not an exhaustive list of ideas, examples include:
 - cloning a template repo
 - caching assignment dependencies
 - pre-compiling and linking code

Leverage the `static-docker-commands` and `static-file-ops` to reduce post submission work for the grader.

## Unexpected Grading Results

When graders work locally but unexpectedly give the wrong output during docker grading,
the easiest way to pinpoint issues is to manually step through the grading process inside the grading container.

The assignment grader image will likely follow the naming scheme of `autograder.<course-name>.<assignment-name>`.
To confirm the exact name of the grading image, list all Docker images by running the `docker images` command.

Using `assignment-01` from `my-first-course` from the [create-a-course tutorial](create-a-course.md) as an example,
run the following (from the base directory (the one with the [README.md](../../README.md) in it)):
```
docker run -it --rm -v docs:/tmp/docs autograder.my-first-course.assignment-01 sh
```

Let's look at the various parts of this command a bit to better understand it:
 - `docker run`
   - Normal docker run command.
 - `-it`
   - Run the container in "interactive" mode.
 - `--rm`
   - Remove the container after it is closed.
 - `-v docs:/autograder/input`
   - Mount the `docs` directory inside our Docker container.
     This allows us to access our test submissions (or other necessary file) inside of the container.
 - `autograder.my-first-course.assignment-01`
   - The name of the container for the assignment.
 - `sh`
   - Overwrite the CMD of the grader container to open a shell (so you can debug).

Once inside the container, look for anomalies in the set up by doing the following:
 - Inspect file structure of the container
   - Verify that all files are in their expected location.
 - Review file permissions
   - Make sure the grader has proper access to necessary files and executables.

If files are in the correct locations with the correct permissions,
proceed by stepping through each step of the grader.
 - Perform `post-submission-file-ops` manually
   - Note you will likely have to update relative paths depending on your mount
 - Run the grading script from the correct working directory (`autograder/work` by default).

Most issues with docker set-ups can be discovered by following these steps.
