# Types

This document describes different key types that you may encounter in this project
(usually in the form of a [JSON](https://en.wikipedia.org/wiki/JSON) representation).
All JSON parsing is done using strict JSON standards,
so you cannot have things like trailing commas or comments.
Missing non-required fields will supplied with a sensible default value.
Extra keys will generally be ignored.

 - [Semantic Types](#semantic-types)
   - [Identifier](#identifier)
   - [Course User Reference (CourseUserReference)](#course-user-reference-courseuserreference)
   - [Server User Reference (ServerUserReference)](#server-user-reference-serveruserreference)
   - [Email](#email)
   - [Timestamp](#timestamp)
   - [Pointer](#pointer)
   - [Regex](#regex)
 - [Course](#course)
 - [Assignment](#assignment)
   - [Assignment Grading Images](#assignment-grading-images)
   - [Grader Output](#grader-output-graderoutput)
   - [Test Submission](#test-submission)
   - [Assignments and the LMS](#assignments-and-the-lms)
   - [Analysis Options (AnalysisOptions)](#analysis-options-analysisoptions)
 - [Roles](#roles)
   - [Server Roles (ServerRole)](#server-roles-serverrole)
   - [Course Roles (CourseRole)](#course-roles-courserole)
 - [Tasks (Task)](#tasks-task)
   - [Course Backup Task](#course-backup-task)
   - [Course Email Logs Task](#course-email-logs-task)
   - [Course Report Task](#course-report-task)
   - [Course Scoring Upload Task](#course-scoring-upload-task)
   - [Course Update Task](#course-update-task)
 - [Scheduled Time (ScheduledTime)](#scheduled-time-scheduledtime)
   - [every - Duration Specification (DurationSpec)](#every---duration-specification-durationspec)
   - [daily - Time of Day Specification (TimeOfDaySpec)](#daily---time-of-day-specification-timeofdayspec)
 - [Logging](#logging)
   - [Level (LogLevel)](#level-loglevel)
   - [Log Query (LogQuery)](#log-query-logquery)
 - [LMS Adapter (LMSAdapter)](#lms-adapter-lmsadapter)
 - [Late Policy (LatePolicy)](#late-policy-latepolicy)
   - [Baseline Late Policy (baseline)](#baseline-late-policy-baseline)
   - [Constant Penalty Late Policy (constant-penalty)](#constant-penalty-late-policy-constant-penalty)
   - [Percentage Penalty Late Policy (percentage-penalty)](#percentage-penalty-late-policy-percentage-penalty)
   - [Late Days Late Policy (late-days)](#late-days-late-policy-late-days)
 - [Submission Limit (SubmissionLimit)](#submission-limit-submissionlimit)
   - [Submission Limit Window (SubmissionLimitWindow)](#submission-limit-window-submissionlimitwindow)
 - [File Specification (FileSpec)](#file-specification-filespec)
   - [FileSpec -- Path](#filespec----path)
   - [FileSpec -- URL](#filespec----url)
   - [FileSpec -- Git](#filespec----git)
 - [File Operation (FileOp)](#file-operation-fileop)

## Semantic Types

This section discusses types that have simple underlying types (Integer, String, etc),
but adds a layer of semantics on top of it.
These are usually used as part of larger types.

### Identifier

Underlying Type: String

An Identifier is a String used to identify specific entities (like courses or assignments).
Identifiers must only have letters, digits, periods, underscores, and hyphens.
The non-alphanumeric characters cannot be repeated in a sequence (e.g. you can have two periods, just not in a row).
Identifiers must start and end with alphanumeric characters.
Identifiers are case insensitive (they are always stored in lower case).

### Course User Reference (CourseUserReference)

Underlying Type: String

A `CourseUserReference` can be used to generalize targeting any number of course users
(e.g., email recipients).
The following values are allowed:

 - Email - The email of the requested user.
   - Emails that are not enrolled in the course are treated differently depending on the operation
     (e.g., ignored, used normally, result in an error, etc).
 - "\*" - Represents requesting all users in the course.
 - [Course Role](#course-roles-courserole) (e.g., "student", "grader", etc) - Represents requesting all course users with that role.
 - Negative Email - An email address preceded by a minus sign (e.g., "-alice@test.edulinq.org")
   will remove this user from the request (even if they are not currently there).
   This can be useful when using course roles but you want to exclude someone.
 - Negative Course Role - A course role preceded by a minus sign (e.g., "-student")
   will remove all course users with that role from the request.
   This can be useful when using the "\*" but you want to exclude a role.

### Server User Reference (ServerUserReference)

Underlying Type: String

A `ServerUserReference` can be used to generalize targeting any number of server users
(e.g., email recipients).
The following values are allowed:

 - Email - The email of the requested user.
   - Emails that are not present on the server are treated differently depending on the operation
     (e.g., ignored, used normally, result in an error, etc).
 - "\*" - Represents requesting all users on the server.
 - [Server Role](#server-roles-serverrole) (e.g., "user", "creator", etc) - Represents requesting all server users with that role.
 - "\<course id\>::\<course role\>" (e.g., "course101::student") - Represents requesting all users in the target course with that course role.
 - "\*::\<course role\>" (e.g., "\*::student") - Represents requesting all users with that course role in any course.
 - "\<course id\>::\*" - Represents requesting all users in the target course.
 - "\*::\*" - Represents requesting all users enrolled in at least one course.
 - Negative targeting - Exclude the user or group from the request.
   - Any of the previous options preceded by a minus sign
     (e.g., "-alice@test.edulinq.org", "-user", "-*::student") - Represents exluding that user or group from the request.
   - Exclusions can be useful when using broad targeting (e.g., "\*"), but you want to exclude a more specific targeting group (e.g., server role, course, course role, email, etc).

### Email

Underlying Type: String

An email address.
Email addresses are usually used to identify a user on the server.
Some places that accept an email address use [CourseUserReference](#course-user-reference-courseuserreference).

### Timestamp

Underlying Type: Integer (64 bit signed)

Timestamps represent a specific instance in time (a datetime).
The numeric value is UNIX millisecond time (the number of milliseconds since the [UNIX epoch](https://en.wikipedia.org/wiki/Unix_time)).
Note that the UNIX epoch is in UTC, so this type always denotes the same timezone (UTC).

In cases where strings are converted to a timestamp (and that string is not all digits) and no timezone information is supplied,
then the server's local time will be used.

### Pointer

Underlying Type: \<variable\>

JSON does not explicitly support pointers, but Go does.
Therefore, many Autograder types may be represented by a pointer.
In documentation, this will be denoted with a star (\*) before another type name (e.g., `*Integer`).

In JSON, a field with a pointer type can take either the base value or `null`.
There is no need to use any addressing or referencing operators (e.g., `&`) in your JSON, the autograder will handle that automatically.
For example, a `*Integer` field can take a normal Integer (e.g. `123`) or a `null` value.

### Regex

Underlying Type: String

A [regular expression](https://en.wikipedia.org/wiki/Regular_expression).
All regular expressions will use the standard regular expression [library](https://pkg.go.dev/regexp) in Go,
which uses the [RE2 Regular Expression Syntax](https://github.com/google/re2/wiki/Syntax).

## Course

A course is the core organizational unit in the autograder.
Users may be enrolled in as many courses as available, but can have only one [role](#course-roles-courseRole) within that course.

A course in the autograder is mainly comprised of:
 - An Identifier
 - Users (with various [roles](#course-roles-courseRole)
 - [Assignments](#assignment)
 - [Tasks](#tasks-task)

| Name               | Type               | Required | Description |
|--------------------|--------------------|----------|-------------|
| `id`               | Identifier         | true     | Identifier for an course. Must be unique within a server. It is recommended to add term and section information were applicable, e.g., `course101-fall24-section01` or `course101-f24-s01`. |
| `name`             | String             | false    | Display name for an course. Defaults to the course's Identifier. |
| `late-policy`      | \*LatePolicy       | false    | The default late policy to use for all assignments in this course. |
| `submission-limit` | \*SubmissionLimit  | false    | The default submission limit to enforce for all assignments in this course. |
| `source`           | \*FileSpec         | false    | The canonical source for a course. This should point to where the autograder can fetch the most up-to-date version of this course. |
| `lms`              | \*LMSAdapter       | false    | Information about how this course can interact with its Learning Management System (LMS). |
| `tasks`            | List[Task]         | false    | Specifications for tasks to run. |

Depending on your LMS, you may also think of an autograder course as a "section",
or specific instantiation of a course in a term.

## Assignment

Assignments represent common gradable activities in a course.
The fields of an assignment are as follows:

| Name               | Type               | Required | Description |
|--------------------|--------------------|----------|-------------|
| `id`               | Identifier         | true     | Identifier for an assignment. Must be unique within a server. |
| `name`             | String             | false    | Display name for an assignment. Defaults to the assignment's Identifier. |
| `sort-id`          | String             | false    | An optional ID to use when sorting assignments. If not provided, an assignment's id will be used when ordering is required. |
| `due-date`         | \*Timestamp        | false    | The due data for an assignment. This can be synced from the course LMS. |
| `max-points`       | float              | false    | The maximum number of points available for the assignment. Although not required when grading, some late policies need this. |
| `lms-id`           | String             | false    | The LMS Identifier for this assignment. May be synced with the LMS if the assignment's name matches. |
| `late-policy`      | \*LatePolicy       | false    | The late policy to use for this assignment. Overrides any late policy set on the course level. |
| `submission-limit` | \*SubmissionLimit  | false    | The submission limit to enforce for this assignment. Overrides any limits set on the course level. |
| `max-runtime-secs` | Integer            | false    | The maximum number of sections a grader is allowed to run before being killed (cannot be greater than system limit set by `docker.runtime.max` config option. |
| `analysis-options` | AnalysisOptions    | false    | Options for code analysis. |
| `image`            | String             | true     | The base Docker image to use for this assignment. |
| `pre-static-docker-commands`  | List[String]   | false | A list of Docker commands to run before static files are copied into the image. |
| `post-static-docker-commands` | List[String]   | false | A list of Docker commands to run after static files are copied into the image. |
| `invocation`                  | List[String]   | false | The command to run when the grading container is launched (Docker's `CMD`). If not set, the image's default `ENTRYPOINT/CMD` is used. |
| `static-files`                | List[FileSpec] | false | A list of files to copy into the image's `/work` directory. |
| `pre-static-file-ops`        | List[FileOp]   | false | A list of file operations to run before static files are copied into the image. |
| `post-static-file-ops`       | List[FileOp]   | false | A list of file operations to run after static files are copied into the image. |
| `post-submission-file-ops`   | List[FileOp]   | false | A list of file operations to run after a student's code is available in the `/input` directory. |

Note that there are few required fields.

### Assignment Grading Images

For each assignment, a [Docker image](https://www.docker.com/) is constructed using the assignment's configuration.
This image is then invoked each time a student's submission needs to be graded,
creating an identical and isolated grading environment for each student.

Images will be created with some default directories:
 - `/autograder` -- The base directory for autograder materials. When running commands, this is the default directory.
 - `/autograder/input` -- Where student code is places. This directory is read-only.
 - `/autograder/output` -- Where the [output of grading](#grader-output) (specifically `/autograder/output/results.json`) should be written.
 - `/autograder/work` -- The "working" directory for grading. All static files are copied here.
 - `/autograder/scripts` -- Where miscellaneous scripts are located.

The general grading workflow is as follows:
 - Student makes a submission request to the autograder.
 - The autograder copies the student's code to a temporary location.
 - The assignment image is invoked with two mounts/volumes:
    - `/autograder/input` (read-only) which contains the student's submission.
    - `/autograder/output` where the grader is expected to write the results of grading (`/autograder/output/results.json`).
 - After the grading container stops running, the autograder collects the grading artifacts (results.json, stdout, and stderr), and inserts them into the database.
 - The autograder now responds to the student's submission request with a summary of the results.

Building an assignment image can seem daunting, so here we will run through the different options in the order they are executed.

`image`  
The first component of an assignment image is the base Docker image.
This can be any image specification that can appear in a [Docker FROM statement](https://docs.docker.com/reference/dockerfile/#from).
We recommend that you specify a specific tag version or digest so you can be sure the assignment is using the correct image.
This project maintains [some general images](https://github.com/edulinq/autograder-docker) that you can use,
or you can create your own custom images.

`pre-static-docker-commands`  
This is a list of commands that we will blindly copy into the assignment's Docker file before static files are copied.
This is a good opportunity to install system dependencies in your image.
These commands do not just include POSIX commands that you would use with [RUN](https://docs.docker.com/reference/dockerfile/#run), but any Dockerfile commands like
[ADD](https://docs.docker.com/reference/dockerfile/#add),
[ENV](https://docs.docker.com/reference/dockerfile/#env),
or [USER](https://docs.docker.com/reference/dockerfile/#user).
Note that we cannot guarantee that the commands you add here will not break the image (so you should test locally).
The `build.keep` config option is useful for keeping around Docker build directories (contexts) for manual inspection.

`pre-static-file-ops`  
These [FileOps](#file-operation-fileop) are run inside the Docker context (build directory) before copying in static files.
Therefore, these commands are not in the Docker image itself, but can modify the files that will then go into the image.
As per standard Docker building rules, only files inside the Docker context can be accessed.
These files operations are generally not useful for most common cases,
the `post-static-file-ops` tends to be more useful.

`static-files`  
These files (FileSpecs) will be copied into the assignment image's `/autograder/work` directory.
All path-based FileSpecs must be relative paths,
and they are all relative to the directory the `assignment.json` is located in.

`post-static-file-ops`  
These [FileOps](#file-operation-fileop) are run inside the Docker context (build directory) before copying in static files.
Therefore, these commands are not in the Docker image itself, but can modify the files that will then go into the image.
As per standard Docker building rules, only files inside the Docker context can be accessed.
Note that static files are copied into the `work` directory, and this command is executed inside the parent of the `work` directory.
So if you want to access static files, you will have to path inside the `work` directory.

`post-static-docker-commands`  
These commands are added to the Dockerfile after the static files are copied.
This is a good opportunity to move around your static files to their preferred location.

`post-submission-file-ops`  
Post submission file operations are intended to be run after the student has submitted their code (and that code is available in the `/autograder/input` directory).
When the assignment image is created, the post submission file operations are written to `/autograder/scripts/post-submission-ops.sh`.
It is the responsibility of the assignment image to execute this file before starting the grader.
The [provided images](https://github.com/edulinq/autograder-docker) already do that,
but you will have to ensure it is done if you create any custom images.

`invocation`  
The invocation is the command to run your actual grader (the Docker image's [CMD](https://docs.docker.com/reference/dockerfile/#cmd)).
We recommend that you wrap your grader in a shell script for easy control and invocation.
The invocation will take place (by default) in the `/autograder` directory.

### Grader Output (GraderOutput)

When a grader finishes running, it is supposed to create a JSON file (`/autograder/output/results.json`)
that describes the result of grader.
The fields for this file (the `GraderOutput` type) are as follows:
| Name                 | Type                 | Required | Description |
|----------------------|----------------------|----------|-------------|
| `name`               | String               | true     | The name of the assignment. This is used as a display name when formatting output for students. |
| `questions`          | List[GradedQuestion] | true     | The result of grading each question. |
| `grading_start_time` | Timestamp            | false    | The time grading started for this assignment. Will default to when the autograder attempted to start the grading container. |
| `grading_end_time`   | Timestamp            | false    | The time grading ended for this assignment. Will default to when the grading container finishes. |
| `prologue`           | String               | false    | Optional text to include at the beginning of a grading report. |
| `epilogue`           | String               | false    | Optional text to include at the end of a grading report. |

Each question (`GradedQuestion`) has the following fields:
| Name                 | Type                 | Required | Description |
|----------------------|----------------------|----------|-------------|
| `name`               | String               | true     | The display name for the question. |
| `max_points`         | Float                | true     | The maximum score possible (not including extra credit) for this question. |
| `score`              | Float                | true     | The score this submission received on this question. |
| `message`            | String               | false    | Optional grading notes to send the student. This is where feedback should be sent to students about missed points. |
| `grading_start_time` | Timestamp            | false    | The time grading started for this question. |
| `grading_end_time`   | Timestamp            | false    | The time grading ended for this question. |

Note that all grading output will be visible to the student who made the submissions.
So, it should not contain any information about grading that students should not see (like inputs to hidden test cases).

### Test Submission

A test submission is a directory that contains a sample submission (code) along with the expected output of the grader.
Like courses and assignments, a test submission is identified by a specific file: `test-submission.json`.
All other files/directories in that parent directory are considered part of the submission.
The `test-submission.json` has the following fields:
| Name              | Type         | Required | Description |
|-------------------|--------------|----------|-------------|
| `ignore_messages` | Boolean      | false    | Ignore the `messages` field when comparing expected output. Defaults to `false`. |
| `result`          | GraderOutput | true     | The expected output to compare against. |

Test submissions are very useful to have for assignments/graders.
They give anyone working with the course reference code to look at,
they provide a simple way to test your grader,
they provide a pre-packaged submission to the autograder for testing autograder commands.
We suggest that you have at least three test submissions for each assignment:
 - A default implementation that only contains the base code provided to students (an empty implementation).
 - A full solution that is considered the gold-standard.
 - A solution that does not compile/parse. This allows you to test if you grader can handle broken code.

### Assignments and the LMS

Certain assignment information can be synced to the autograder from the course's LMS.
Information that can be synced:
 - `name` (see below)
 - `lms-id` (see below)
 - `due-date`
 - `max-points`

Before information is synced over, an autograder assignment must match up to an LMS assignment.
The most direct way to match is to populate the `lms-id` field with the LMS identifier for the assignment.
However, this would require updating the `lmd-id` field for each section/term.
You can also ensure that the assignment names in the autograder and LMS are the same (and there are no other assignments with the same name).
On a full name match, then autograder will sync over the `lms-id` from the course's LMS.

### Analysis Options (AnalysisOptions)

The analysis options type allows options to be passed to code analysis for assignments.
It has the following fields:
| Name               | Type            | Required | Description |
|--------------------|-----------------|----------|-------------|
| `include-patterns` | List[Regex]     | false    | Any source file eligible for code analysis must match at least one of these patterns. When not specified or empty, ".+" will be used (which will match any non-empty value). |
| `exclude-patterns` | List[Regex]     | false    | Any source file that matches any of these patterns will not be used in code analysis. |
| `template-files`   | List[FileSpec]  | false    | A list of files to use as "templates" during pairwise analysis. Similarity engines can try to ignore template/boilerplate code when computing similarities. Any file paths must be local (relative). |
| `template-file-ops`| List[FileOp]    | false    | A list of file operations to transform the template files with. |
| `engine-options`   | Map             | false    | A map keyed by engine name to it's respective engine options. |

`engine-options` contains `OptionsMap`, a key–value map of type `map[string]any` holding options specific to that engine.

During a pairwise code analysis,
the options of the assignment for the submission with the [lexicographically](https://en.wikipedia.org/wiki/Lexicographic_order) smaller id will always be used.

Template files allow you to specify code that pairwise similarity engines should ignore when computing code similarity (e.g. boilerplate code).
To align template files with files that students submit,
the [relative path](https://en.wikipedia.org/wiki/Path_(computing)#Absolute_and_relative_paths) ob both files must match exactly.
This means that if students are submitting files and not directories then just the filenames need to match,
if students are submitting directories then the template files must be in directories with the same name.
When fetching and setting up the template files, the file specs (`template-files`) will be fetched from the assignment's base directory,
and the file operations (`template-file-ops`) will be executed in a directory that only has the fetched template files in it.
This means that you should use the `template-files` to collect the files you need,
and `template-file-ops` to move around and adjust those files so they properly match student submissions.

#### Include/Exclude Patterns

The include/exclude patterns decide which source files will be included in code analysis for each assignment.
When working with these patterns, keep the following in mind:

1. The include/exclude patterns are not [globs](https://en.wikipedia.org/wiki/Glob_(programming)),
    like you may find in a shell/command-line.
    Instead, they are full [Regular Expressions](#regex).
2. The patterns operate not on just the filename, but the entire relative path of the submission file.
    So, if a student submits the files [`README.md`, `src/code.c`], then you could use the pattern `src/.+` to match everything in the `src` directory.
3. Inclusions are processed before exclusions.
    So an exclusion pattern of `.*` will exclude everything, regardless of any inclusion patterns.
4. Some source files may be transformed and renamed for analysis.
    The inclusion/exclusion patterns apply after renaming and transformation.
    For example, [iPython Notebooks](https://en.wikipedia.org/wiki/Project_Jupyter#Documents) with the `.ipynb` extensions
    will have their code Python extracted and renamed to `.py`.

#### Engine Options

The engine option type allows engine specific options to be passed for code analysis for assignments.
It is a map of engine name to its respective engine options.

Example Structure: 
```json
{
    "jplag": {
        "min-tokens": 7
    },
    "dolos": {
        "kgram-length": 12
    }
}
```

Currently, we have support for two code analysis tools for which engine options are used. 
They are [JPlag](https://github.com/jplag/JPlag/wiki) and [Dolos](https://dolos.ugent.be/docs/). 

##### JPlag 

JPlag is a code analysis tool that compares code files to identify for any suspicious similarities. 
More info on it can be found here: [JPlag](https://github.com/jplag/JPlag/wiki)
The values from its `OptionsMap` in `engine-options` are parsed into CLI arguments for the JPlag container.

Current Supported Options:
| Name               | Type           | Required | Description |
|--------------------|----------------|----------|-------------|
| `min-tokens`       | Integer        | false    | The minimum number of consecutive tokens that must match between two code submissions to be considered a plagiarism match. |

Example Structure:
``` json
"jplag": {
    "min-tokens": 7
}
```

#### Dolos

Dolos is a code analysis tool optimized for performance and detailed similarity reporting. 
More on dolos can be found here: [Dolos](https://dolos.ugent.be/docs/).
The values from its `OptionsMap` in `engine-options` are parsed into CLI arguments for the Dolos container.

Current Supported Options:
| Name               | Type           | Required | Description |
|--------------------|----------------|----------|-------------|
| `kgram-length`     | Integer        | false    | Sets the minimum number of tokens in a k-gram. Common fragments between two files that are shorter than `$$k$$` tokens will not be found during plagiarism detection. |
| `kgrams-in-window` | Integer        | false    | The size of the window used during winnowing algorithm. It select one k-grams from each overlapping window of `w` subsequent k-grams. |

Example Structure: 
```json
"dolos": {
    "kgram-length": 12,
    "kgrams-in-window": 23
}
```

## Roles

Roles are used to define privileges for a user within the server and each course.
All users will have a single "server role" that defines their abilities on the server,
as well as a "course role" for each course they are enrolled in.
Roles are externally represented by case-insensitive strings.

In general, a higher role is always allowed when a lower role is required.
For example, an API endpoint may say that is requires a server admin,
but a server owner (which has more privilege than an admin) will always suffice.
Imaging that the phrase "or greater" is always omitted.

### Server Roles (ServerRole)

Server roles determine a user's abilities on the server.
Every user has exactly one server role.
Roles further down the list are considered more privileged.

| Role      | Description |
|-----------|-------------|
| `user`    | The default role for server users. |
| `creator` | Creators are allowed to create courses (which they will own), but are otherwise the same as users. |
| `admin`   | Server administrators. Generally allowed to do any operation on the server. |
| `owner`   | Server owners. |

Users with a role of admin are free to execute any operation that require course-specific permissions.

### Course Roles (CourseRole)

Course roles determine a user's abilities within a specific course.
A user enrolled in a course will have exactly one course role associated with that course.
Roles further down the list are considered more privileged.

| Role      | Description |
|-----------|-------------|
| `other`   | Course members who are not allowed to make submissions, but otherwise observe the course. |
| `student` | Standard course members who submit assignments and can view only their own work. |
| `grader`  | Course members who have access to submissions/grades for all students. These users have full access to the **content** of the course, but **cannot administer** it (e.g. they can see all the users but cannot edit them). |
| `admin`   | Course members who can administrate the course. These users have full access to the content of the course and can administer it. |
| `owner`   | Owners of the course. |

## Tasks (Task)

Tasks are asynchronous processes that run on the server (usually related to a course).
Tasks share the common structure below:

| Name       | Type          | Required | Description |
|------------|---------------|----------|-------------|
| `type`     | String        | true     | The type of task. All allowed types are discussed in this section. |
| `name`     | String        | false    | A display name for this task. |
| `disabled` | Boolean       | false    | Set to true to disable this task from running. |
| `when`     | ScheduledTime | true     | When to run this task. |
| `options`  | Map           | false    | Options used by the specific type of task. |

The `type` of the task determines what values will be looked for in `options`.
The available task types will be discussed in the rest of this section.

### Course Backup Task

A backup task backs up the course information to the server's backup location.

Type: `backup`

No additional options.

Basic Example:
```json
{
    ... the rest of a course object ...
    "tasks": [
        {
            "type": "backup",
            "when": {
                "daily": "3:00"
            }
        }
    ]
}
```

### Course Email Logs Task

The email logs task sends an email to the target users containing matching logs.

Type: `email-logs`

Additional Options:
| Name         | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `to`         | List[CourseEmailSpec] | true     | A list of emails to send the results to. At least one recipient must be listed. |
| `send-empty` | Boolean               | false    | If true, the email will still be sent even if no query results were found. |
| `query`      | LogQuery              | true     | The log query to execute. |

Basic Example:
```json
{
    ... the rest of a course object ...
    "tasks": [
        {
            "type": "email-logs",
            "when": {
                "daily": "3:00"
            },
            "options": {
                "to": [
                    "admin"
                ],
                "query": {
                    "past": "24h",
                    "level": "ERROR"
                }
            }
        }
    ]
}
```

### Course Report Task

The report task sends an email to the target users summarizing the current submissions for each assignment.

Type: `report`

Additional Options:
| Name         | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `to`         | List[CourseEmailSpec] | true     | A list of emails to send the report to. At least one recipient must be listed. |

Basic Example:
```json
{
    ... the rest of a course object ...
    "tasks": [
        {
            "type": "report",
            "when": {
                "daily": "3:00"
            },
            "options": {
                "to": [
                    "owner",
                    "admin",
                    "alice@test.edulinq.org",
                ]
            }
        }
    ]
}
```

### Course Scoring Upload Task

A scoring upload task performs a full scoring
(including reexamining already scored/due assignments) and uploads to scores to the course's LMS.
This task will also recompute late information in the case that any factors have changed
(e.g., a late submission was removed, a due date changed, or a student was given additional late days).
This task has no additional configuration options.

Type: `scoring-upload`

No additional options.

Basic Example:
```json
{
    ... the rest of a course object ...
    "tasks": [
        {
            "type": "scoring-upload",
            "when": {
                "daily": "3:00"
            }
        }
    ]
}
```

### Course Update Task

Course update tasks will update a course from source and perform the standard update protocol
(syncing with the LMS, build assignment images, etc).

Type: `update`

No additional options.

Basic Example:
```json
{
    ... the rest of a course object ...
    "tasks": [
        {
            "type": "update",
            "when": {
                "daily": "3:00"
            }
        }
    ]
}
```

## Scheduled Time (ScheduledTime)

A `ScheduedTime` describes when to run some procedure (usually a [Task](#tasks-task)).
It has two exclusive fields that allow for different ways of describing run times: `every` and `daily`.
One and only one of the two fields must be populated.

| Name    | Type          | Required | Description |
|---------|---------------|----------|-------------|
| `every` | DurationSpec  | false    | Specifies the period between procedure runs. |
| `daily` | TimeOfDaySpec | false    | Specifies when a procedure should run each day. |

### every - Duration Specification (DurationSpec)

A DurationSpec allows you to specify the period between procedure runs, e.g., "every 4 hours".
This type has several fields that represent time units.

| Name      | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `days`    | Integer | false    | Days        |
| `hours`   | Integer | false    | Hours       |
| `minutes` | Integer | false    | Minutes     |
| `seconds` | Integer | false    | Seconds     |

Here are some general rules for a DurationSpec are:
 - No negative values.
 - At least one field must be populated.
 - Total time must not 100 days.

**Examples**

"every 4 hours"
```json
{
    "when": [{
        "every": {
            "hours": 4
        }
    }]
}
```

"every 1 day, 2 hours, 3 minutes, and 4 seconds"
```json
{
    "when": [{
        "every": {
            "days": 1,
            "hours": 2,
            "minutes": 3,
            "seconds": 4
        }
    }]
}
```

"(every 4 hours) and (every 7 days)"
```json
{
    "when": [
        {
            "every": {
                "hours": 4
            }
        },
        {
            "every": {
                "days": 7
            }
        }
    ]
}
```

### daily - Time of Day Specification (TimeOfDaySpec)

A TimeOfDaySpec allows you to specify when something will run each day.
Using this implies that you want the procedure to run once a day, and you are selecting when to run it each day.

A TimeOfDaySpec is just a string formatted either as:
 - `HH:MM`
 - `HH:MM:SS`

All times MUST be formatted in [24-hour time](https://en.wikipedia.org/wiki/24-hour_clock).
The timezone of the server will be used to interpret this time.

**Examples**

"every day at 23:59 (11:59 PM)"
```json
{
    "when": [{
        "daily": "23:59"
    }]
}
```

"every day at 12:34:56"
```json
{
    "when": [{
        "daily": "12:34:56"
    }]
}
```

"(every 4 hours) and (every day at 23:59)"
```json
{
    "when": [
        {
            "every": {
                "hours": 4
            }
        },
        {
            "daily": "23:59"
        }
    ]
}
```

## Logging

Logging is a core functionality of the autograder and there are some relevant types a user should be aware of.

### Level (LogLevel)

As with most logging infrastructure, logs are assigned a level.
Server administrators can set the level that they wish to record to stderr and the database respectively
using the `log.text.level` and `log.backend.level` options.
When a level is set, all logs at or above that level are included.
Log levels are one of the following strings.

| Level   | Description |
|---------|-------------|
| `TRACE` | The lowest level available. Used for fine-grained detail (like loop counters) while debugging. |
| `DEBUG` | A logging level to use when debugging. |
| `INFO`  | The default logging level. Includes messages that those running the server may find useful. |
| `WARN`  | Warnings that do not stop the server from running, but are good to know about. |
| `ERROR` | Errors that do not crash the server, but imped certain functionality. Errors should be addressed as soon as possible. |
| `FATAL` | Critical errors that the server cannot recover from. Should only occur as soon as the server starts, not during prolonged running. |
| `OFF`   | Turns off all logging. |

The default level is `INFO`.

### Log Query (LogQuery)

A log query is a structured object that can be used to retrieve relevant log records.
A query returns all records where all non-empty fields match.

| Name                | Type       | Required | Description |
|---------------------|------------|----------|-------------|
| `level`             | LogLevel   | false    | Matches records at or above this log level. Defaults to "INFO" |
| `after`             | String     | false    | Matches records after this time. The string should formatted either as [RFC 3339](https://en.wikipedia.org/wiki/ISO_8601#RFCs) or an integer ([milliseconds since UNIX epoch](https://en.wikipedia.org/wiki/Unix_time)). |
| `past`              | String     | false    | Matches records that have been logged within this duration (e.g., "in the past hour"). Must have the pattern \<int\>\<unit\> where the units may be "s" (seconds), "m" (minutes), or "h" (hours). For example: "2h" for two hours. |
| `target-course`     | Identifier | false    | Matches records about the given course. |
| `target-assignment` | Identifier | false    | Matches records about the given assignment. If present, the query must also have `target-course` populated. |
| `target-email`      | Email      | false    | Matches records about the given user/email. |

The contents of a query determine the required permissions for that query.
The general rules are:
 - Sever admins can make any query.
 - A user can always query for logs about themselves.
 - A course admin can always query for logs about their course.
 - If not timezone is specified in `after` or `past`, the server's local timezone is used.

## LMS Adapter (LMSAdapter)

A [Leaning Management System](https://en.wikipedia.org/wiki/Learning_management_system) (LMS) is a system that is used to administer courses and track student grades.
The most popular are [Canvas](https://en.wikipedia.org/wiki/Instructure), [Blackboard](https://en.wikipedia.org/wiki/Blackboard_Learn), and [Moodle](https://en.wikipedia.org/wiki/Moodle).

An LMS adapter contains all the information necessary to link an autograder course with an LMS course.

| Name                   | Type       | Required | Description |
|------------------------|------------|----------|-------------|
| `type`                 | String     | true     | The type of the LMS being connected to. Currently the only valid value is "canvas". |
| `base-url`             | String     | true     | The base URL of the LMS instance the course lives on, e.g. "https://canvas.university.edu". |
| `course-id`            | String     | true     | The course identifier within the LMS. (This is not the autograder course id.) |
| `api-token`            | String     | false    | The token used to authenticate API requests to the LMS. |
| `sync-user-attributes` | Boolean    | false    | Sync attributes of users (e.g. name) when syncing users between the autograder and LMS. |
| `sync-user-adds`       | Boolean    | false    | Sync new users when syncing users between the autograder and LMS. |
| `sync-user-removes`    | Boolean    | false    | Sync removed users when syncing users between the autograder and LMS. Note that this can cause issues if you have manually added users that do not appear in your LMS. |
| `sync-assignments`     | Boolean    | false    | Try to sync assignment details (name, due date, etc) when syncing with the LMS. |

## Late Policy (LatePolicy)

The autograder can apply one of several late policies to an assignment.
The late policy for an assignment may be set in either the course or assignment configs.
When set in the course config, assignments may override the behavior for a specific assignment by specifying their own policy.

All late policies share common fields:

| Name                   | Type       | Required | Description |
|------------------------|------------|----------|-------------|
| `type`                 | String     | true     | The type of late policy being used. Valid values are: `baseline`, `constant-penalty`, `percentage-penalty`, and `late-days`. |
| `reject-after-days`    | Integer    | false    | After this number of days past the assignment due date, do not accept any more submissions. Submissions past this time are fully ignored, they will not be neither stored nor stored. A zero (or no) value will result in submissions never being rejected for being late. |

### Baseline Late Policy (baseline)

The `baseline` policy will not remove points for tardiness, but will still reject submissions that are too late.

### Constant Penalty Late Policy (constant-penalty)

The `constant-penalty` late policy will apply a constant penalty for each day a submission is late.

| Name      | Type  | Required | Description |
|-----------|-------|----------|-------------|
| `penalty` | Float | true     | The penalty to apply for each day of being late. Must be greater than zero. |

### Percentage Penalty Late Policy (percentage-penalty)

The `percentage-penalty` late policy will deduct a percentage of an assignment's max points for each day a submission is late.

| Name      | Type  | Required | Description |
|-----------|-------|----------|-------------|
| `penalty` | Float | true     | The proportion of the assignment's max points to apply as a penalty for each day of being late. Must be in larger than 0.0 and less than or equal to 1.0. |

### Late Days Late Policy (late-days)

The `late-days` late policy is a more complicated late policy that allows each student to use their own bank of late or grace days.
Below is a textual description (which may be hard to follow).
Below the textual description are examples that may be easier to follow.

At the start of the term, each student is given a specific amount of late days (3-5 is common).
For each assignment, a student may use a pre-determined maximum number of late days.
When a student turns in an assignment late, first their effective grace days is computed by taking the minimum of
the number of actual days late, the number of remaining grace days the student has available, and the number of maximum late days the assignment allows student's to use.
If the number of effective grace days is not sufficient to cover the number of actual days the assignment is late,
then each remaining late day is penalized in the same way as the `percentage-penalty` late policy.

| Name                 | Type    | Required | Description |
|----------------------|---------|----------|-------------|
| `penalty`            | Float   | true     | The proportion of the assignment's max points to apply as a penalty for each day of being late (after applying grace days). Must be in larger than 0.0 and less than or equal to 1.0. |
| `max-late-days`      | Integer | true     | The maximum number of late days students may expend on this assignment. Must be less than or equal to `reject-after-days`. |
| `late-days-lms-id`   | String  | false*   | The LMS ID for the assignment that will be used to track late days for each student. This assignment is usually not included in the final grade, but serves as a great place for students to view how many late days they have left. The autograder should have permissions to read and write this assignment. Instructors should populate it with the initial number of available late days for each student. |
| `late-days-lms-name` | String  | false*   | The name of the late days assignment in the LMS. This can be used to sync with the late days assignment if `late-days-lms-id` is not provided.

At least one of the `late-days-lms-id` and `late-days-lms-name` fields must be set.

For example:
```json
{
    "type": "late-days",
    "reject-after-days": 4,
    "penalty": 0.10,
    "max-late-days": 2,
    "late-days-lms-id": "ABC123"
}
```

Using the above configuration, let's look at a few examples:
| Student | Days Late | Grace Days Available | Used Late Days | Raw Score | Final Score | Notes |
|---------|-----------|----------------------|----------------|-----------|-------------|-------|
| Alice   | 0         | 5                    | 0              | 100       | 100         | Nothing is used when the assignment is not late. |
| Bob     | 1         | 5                    | 1              | 100       | 100         | Bob used 1 late day, but receives no penalty (since he had enough late days to use). |
| Claire  | 2         | 2                    | 2              | 100       | 100         | With this submission, Claire has now used all her late days. There is no penalty for this submission, but will be next time. |
| Doug    | 3         | 5                    | 2              | 100       | 90          | Although Doug has many late days, the assignment allows a maximum of 2 late days. He will be penalized for 1 day. |
| Emma    | 4         | 2                    | 2              | 100       | 80          | Emma used all their late days and will still need to be penalized for 2 more days. |
| Francis | 5         | 2                    | 0              | ?         | ?           | Francis submitted too late. Their submission has been rejected and will not receive a formal score. No late days will be used. |

## Submission Limit (SubmissionLimit)

Submission limits put a limit on the number or rate of submissions a student can make to an assignment.
There are multiple types of submission, and they call all be used.
A submission is rejected if any of the specified limits trigger.
Submission limits are not enforced for course graders.

| Name               | Type                    | Required | Description |
|--------------------|-------------------------|----------|-------------|
| `max-attempts`     | \*Integer               | false    | This is the total number of max submissions a student is allowed to have. |
| `window`           | \*SubmissionLimitWindow | false    | This specifies a sliding window limiting the number of submissions. |

### Submission Limit Window (SubmissionLimitWindow)

| Name               | Type    | Required | Description |
|--------------------|---------|----------|-------------|
| `allowed-attempts` | Integer | true     | The number of allowed submissions within this window. |
| `duration`         | String  | true     | The size of the window. Must have the pattern \<int\>\<unit\> where the units may be "s" (seconds), "m" (minutes), or "h" (hours). For example: "2h" for two hours. |

## File Specification (FileSpec)

A file specification (FileSpec) defines how to access a specific file (or dir).
At its easiest, it is just the path to a file, but can also specify things like a URL, a Git repository,
or a [glob](https://en.wikipedia.org/wiki/Glob_(programming)) pattern for matching multiple files or directories.
FileSpecs are used for things like specifying the canonical source for a course or the files that an assignment grader requires.

All types of FileSpecs share some common fields:

| Name       | Type   | Required | Description |
|------------|--------|----------|-------------|
| `type`     | String | true     | The type of FileSpec. Must be one of: "path", "git", or "url". |
| `path`     | String | true     | The path to the resource this FileSpec points to. Will take different forms depending on the type. |
| `dest`     | String | false    | The name to refer to the FileSpec once it has been fetched. If not specified, this is inferred from the path (e.g. the name of the file). |
| `username` | String | false    | The username for authentication. |
| `token`    | String | false    | The token/password for authentication. We recommend using tokens with fine-grained read-only access when possible. |

As a special case, some FileSpecs can be parsed from a string that is not JSON.
The most common case of this is just using a path directly.
A string (that does not start with `http://` or `git::`) is interpreted as a path-type FileSpec.

### FileSpec -- Path

A FileSpec with `type` equal to `path` points to an absolute or relative path accessible from the current machine.
`dest` must be a directory if multiple files/directories are being copied, a new directory will be created if `dest` doesn't exist.
When a relative path is specified, additional context is required to know the relative base.
For example, a FileSpec in an assignment config is relative to the assignment directory (the directory where the `assignment.json` file lives).

The path may include a glob pattern to target multiple files/directories.
Glob patterns must follow Go [glob pattern standard](https://pkg.go.dev/path/filepath#Match).

In most cases where a FileSpec is parsed from a string, e.g., most command-line cases, a path FileSpec can be given as a normal path instead of a JSON object.
`type` and `path` will be set properly, and `dest` will be defaulted to the given path's base name.

**Examples**

Here are some examples that all attempt to point to the README for this repository.

Absolute path (assumes this repo is at the root directory):
```json
{
    "type": "path",
    "path": "/autograder-server/README.md"
}
```

Relative path (assumes this directory is the relative base):
```json
{
    "type": "path",
    "path": "../README.md"
}
```

Rename the output file to "instructions.md":
```json
{
    "type": "path",
    "path": "/autograder-server/README.md",
    "dest": "instructions.md"
}
```

Multiple files with a glob:
```json
{
    "type": "path",
    "path": "test_cases/*.c"
}
```

String only:
```json
"/autograder-server/README.md"
```

### FileSpec -- URL

A FileSpec with `type` equal to `url` points to a resource accessible with an HTTP GET request.

Like the path-type FileSpec, a URL FileSpec can be parsed from a string,
as long as the string starts with "http://" or "https://".

**Examples**

Standard URL (note that for GitHub we are pointing to the raw content and not the webpage):
```json
{
    "type": "url",
    "path": "https://raw.githubusercontent.com/edulinq/autograder-server/refs/heads/main/README.md"
}
```

Rename the output file to "instructions.md":
```json
{
    "type": "url",
    "path": "https://raw.githubusercontent.com/edulinq/autograder-server/refs/heads/main/README.md",
    "dest": "instructions.md"
}
```

String only:
```json
"https://raw.githubusercontent.com/edulinq/autograder-server/refs/heads/main/README.md"
```


### FileSpec -- Git

A FileSpec with `type` equal to `git` points to a Git repository accessible with either no credentials or using `username` and `token`.
The `path` field should be a Git URL.
HTTP URLs are preferred  over SSH unless you setup SSH keys on your autograder server ahead of time.

| Name        | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `reference` | String | false    | The name of the git reference to use, e.g., commit hash, branch name, or tag name. Defaults to the repository's default branch. |

When using a Git FileSpec, it is strongly recommended to include a reference.
This allows you to more closely control the exact files you are getting.
For a course source, specifying a non-default branch as a reference allows you to develop without worrying about accidentally pushing changes to the autograder.
For an assignment resource, specifying a commit hash allows to autograder to know if a file has changes and the assignment's Docker container needs to be rebuilt.

Like the path-type FileSpec, a Git FileSpec can be parsed from a string.
The strong should be formatted as: `git::<path>[@<reference>]` (where the reference is options).
See the examples for possible usage.

**Examples**

A simple repository using the default branch without authentication:
```json
{
    "type": "git",
    "path": "https://github.com/edulinq/autograder-server"
}
```

GitHub-style SSH path:
```json
{
    "type": "git",
    "path": "git@github.com:edulinq/autograder-server.git"
}
```

Rename the output directory from "autograder-server" to "ag-server":
```json
{
    "type": "git",
    "path": "https://github.com/edulinq/autograder-server",
    "dest": "ag-server"
}
```

Use a specific branch:
```json
{
    "type": "git",
    "path": "https://github.com/edulinq/autograder-server",
    "reference": "my-cool-branch"
}
```

Use a specific commit hash:
```json
{
    "type": "git",
    "path": "https://github.com/edulinq/autograder-server",
    "reference": "3d5584c1a11307ffdb0ba1c8bb86f40bc36731f4"
}
```

Authenticate against a private repository:
```json
{
    "type": "git",
    "path": "https://github.com/edulinq/autograder-server-secret",
    "username": "secret-name",
    "token": "ghp_abc123"
}
```

GitHub-style embedded credentials:
```json
{
    "type": "git",
    "path": "https://secret-name:ghp_abc123@github.com/edulinq/autograder-server"
}
```

String only without reference:
```json
"git::https://github.com/edulinq/autograder-server"
```

String only with reference:
```json
"git::https://github.com/edulinq/autograder-server@my-cool-branch"
```

## File Operation (FileOp)

A file operation (FileOp) is a description of a simple file operation.
FileOps are always lists of strings (typically two or three strings) representing the operation and its arguments.
The general idea is to support very basic commands that can be easily implemented on any system that may encounter them.

Any path contained within a file operation must obey the following rules:
 - Must be relative.
 - Must not point outside the base/current directory (e.g., `../a.txt`).
 - Must be a POSIX path (use forward slashes as separators).
 - Must not point just to the current directory (`.`).
   The path must point to a path within the current directory (e.g., `a.txt` or even `./a.txt`), not the directory itself.

Below are the currently supported file operations:
| Long Name  | Short Name | Arguments                   | POSIX Equivalent | Description |
|------------|------------|-----------------------------|------------------|-------------|
| `copy`     | `cp`       | source glob path, dest path | `cp -r`          | Copy the source path to the dest path. The source path may be a file or directory. |
| `move`     | `mv`       | source glob path, dest path | `mv`             | Move (or rename) the source path to the dest path. The source path may be a file or directory. |
| `make-dir` | `mkdir`    | path                        | `mkdir -p`       | Create a directory at the given path. The path make already exist. Any required parent directories will be created. |
| `remove`   | `rm`       | glob path                   | `rm -rf`         | Remove the given path. The path may be a file or directory. The path may not exist. |

Any `move`, `copy`, or `remove` command may include [globs](https://en.wikipedia.org/wiki/Glob_%28programming%29) in the source path.
We support the [Go standard for globs](https://pkg.go.dev/path/filepath#Match).
All paths are simplified before globs are expanded, so a glob path like `*/*/..` will be simplified into `*`.

Following the above POSIX equivalents, `move` and `copy` return errors if the source path cannot be found but `remove` will not.

**Examples**

Copy a directory:
```json
    "file-operations": [
        ["copy", "autograder-server", "autograder-server-copy"]
    ]
```

Move/rename a file:
```json
    "file-operations": [
        ["move", "foo.txt", "bar.txt"]
    ]
```
