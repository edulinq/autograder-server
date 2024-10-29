# Types

This document describes different key types that you may encounter in this project
(usually in the form of a [JSON](https://en.wikipedia.org/wiki/JSON) representation).
All JSON parsing is done using strict JSON standards,
so you cannot have things like trailing commas or comments.
Missing non-required fields will supplied with a sensible default value.
Extra keys will generally be ignored.

TODO: ToC

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

### Timestamp

Underlying Type: Integer (64 bit signed)

Timestamps represent a specific instance in time (a datetime).
The numeric value is UNIX millisecond time (the number of milliseconds since the [UNIX epoch](https://en.wikipedia.org/wiki/Unix_time)).
Note that the UNIX epoch is in UTC, so this type always denotes the same timezone (UTC).

### Pointer

Underlying Type: \<variable\>

JSON does not explicitly support pointers, but Go does.
Therefore, many Autograder types may be represented by a pointer.
In documentation, this will be denoted with a star (\*) before another type name (e.g., `*Integer`).

In JSON, a field with a pointer type can take either the base value or `null`.
There is no need to use any addressing or referencing operators (e.g., `&`) in your JSON, the autograder will handle that automatically.
For example, a `*Integer` field can take a normal Integer (e.g. `123`) or a `null` value.

## Course

TOOD

| Name               | Type               | Required | Description |
|--------------------|--------------------|----------|-------------|
| `id`               | Identifier         | true     | Identifier for an course. Must be unique within a server. |
| `name`             | String             | false    | Display name for an course. Defaults to the course's Identifier. |
| `late-policy`      | \*LatePolicy       | false    | The default late policy to use for all assignments in this course. |
| `submission-limit` | \*SubmissionLimit  | false    | The default submission limit to enforce for all assignments in this course. |
| `source`           | \*FileSpec         | false    | The canonical source for a course. This should point to where the autograder can fetch the most up-to-date version of this course. |
| `lms`              | \*LMSAdapter       | false    | Information about how this course can interact with its Learning Management System (LMS). |
| `backup`           | list[\*BackupTask]        | false    | Specifications for tasks to backup the course. |
| `course-update`    | list[\*CourseUpdateTask]  | false    | Specifications for tasks to update this course using its source.
| `report`           | list[\*ReportTask]        | false    | Specifications for tasks to send out a report detailing assignment submissions and scores. |
| `scoring-upload`   | list[\*ScoringUploadTask] | false    | Specifications for tasks to perform a full scoring of all assignments and upload scores to the course's LMS. |
| `email-logs`       | list[\*EmailLogsTask]     | false    | Specifications for tasks to email log entries. |

TODO

## Assignment

JSON

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
| `image`            | String             | true     | The base Docker image to use for this assignment. |
| `pre-static-docker-commands`  | list[String]   | false | A list of Docker commands to run before static files are copied into the image. |
| `post-static-docker-commands` | list[String]   | false | A list of Docker commands to run after static files are copied into the image. |
| `invocation`                  | list[String]   | false | The command to run when the grading container is launched (Docker's `CMD`). If not set, the image's default `ENTRYPOINT/CMD` is used. |
| `static-files`                | list[FileSpec] | false | A list of files to copy into the image's `/work` directory. |
| `pre-static-files-ops`        | list[FileOP]   | false | A list of file operations to run before static files are copied into the image. |
| `post-static-files-ops`       | list[FileOP]   | false | A list of file operations to run after static files are copied into the image. |
| `post-submission-files-ops`   | list[FileOP]   | false | A list of file operations to run after a student's code is available in the `/input` directory. |

LMS sync.
`lms-id` or name match.

TODO

TODO - Move to LMS documentation
Fields synced by LMS: `name`, `due-date`, `max-points`, `lms-id`\*

Build stages
pre-static docker
pre-static file ops
static
post-static file ops
post-static docker
invoation
submission

## Tasks

TODO

### Backup Task (BackupTask)

TODO

### Course Update Task (CourseUpdateTask)

TODO

### Report Task (ReportTask)

TODO

### Scoring Upload Task (ScoringUploadTask)

TODO

### Email Logs Task (EmailLogsTask)

TODO

## LMS Adapter (LMSAdapter)

TODO

## Late Policy (LatePolicy)

TODO

## Submission Limit (SubmissionLimit)

TODO

## File Specification (FileSpec)

TODO

## File Operation (FileOP)
