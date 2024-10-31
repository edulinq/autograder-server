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

### Email

Underlying Type: String

An email address.

When used in the context of a course,
a star ("\*") may be used to reference all users
and a course role ("other", "student", "grader", "admin", or "owner") may be used to include all course members with that role.

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
| `backup`           | List[BackupTask]        | false    | Specifications for tasks to backup the course. |
| `course-update`    | List[CourseUpdateTask]  | false    | Specifications for tasks to update this course using its source.
| `report`           | List[ReportTask]        | false    | Specifications for tasks to send out a report detailing assignment submissions and scores. |
| `scoring-upload`   | List[ScoringUploadTask] | false    | Specifications for tasks to perform a full scoring of all assignments and upload scores to the course's LMS. |
| `email-logs`       | List[EmailLogsTask]     | false    | Specifications for tasks to email log entries. |

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
| `pre-static-docker-commands`  | List[String]   | false | A list of Docker commands to run before static files are copied into the image. |
| `post-static-docker-commands` | List[String]   | false | A list of Docker commands to run after static files are copied into the image. |
| `invocation`                  | List[String]   | false | The command to run when the grading container is launched (Docker's `CMD`). If not set, the image's default `ENTRYPOINT/CMD` is used. |
| `static-files`                | List[FileSpec] | false | A list of files to copy into the image's `/work` directory. |
| `pre-static-files-ops`        | List[FileOP]   | false | A list of file operations to run before static files are copied into the image. |
| `post-static-files-ops`       | List[FileOP]   | false | A list of file operations to run after static files are copied into the image. |
| `post-submission-files-ops`   | List[FileOP]   | false | A list of file operations to run after a student's code is available in the `/input` directory. |

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
invocation
submission

## Roles

TODO

In general, a higher role is always allowed when a lower role is required.
For example, an API endpoint may say that is requires a server admin,
but a server owner (which has more privlege than an admin) will always suffice.
Imaging that the phrase "or greater" is always omitted.

### Server Roles

TODL

### Course Roles

## Tasks

Tasks are asynchronous processes for a course.
Each specific task will have some common fields.

| Name      | Type                | Required | Description |
|-----------|---------------------|----------|-------------|
| `disable` | Boolean             | false    | Set to true to disable this task from running. |
| `when`    | List[ScheduledTime] | false    | When to run this task. Usually contains just one time, but multiple are allowed. |

Tasks are often specified as lists.
This allows you to create multiple instantiations of the same task that have different configurations.
For example, you may want one task to email you WARNING logs every week,
but another task that emails you ERROR logs every day.

### Scheduled Time (ScheduledTime)

A `ScheduedTime` describes when to run a task.
It has two exclusive fields that allow for different ways of describing run times: `every` and `daily`.
One and only one of the two fields must be populated.

| Name    | Type        | Required | Description |
|---------|-------------|----------|-------------|
| `every` | DurationSpec  | false    | Specifies the period between task runs. |
| `daily` | TimeOfDaySpec | false    | Specifies when a task should run each day. |

#### every - Duration Specification (DurationSpec)

A DurationSpec allows you to specify the period between task runs, e.g., "every 4 hours".
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

#### daily - Time of Day Specification (TimeOfDaySpec)

A TimeOfDaySpec allows you to specify when a task should be run each day.
Using this implies that you want the task to run once a day, and you are selecting when to run it each day.

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

### Backup Task (BackupTask)

A backup task backs up the course information to the server's backup location.
This task has no additional configuration options.

Basic Example:
```json
{
    ... the rest of a course object ...
    "backup": [{
        "when": [{
            "daily": "01:00"
        }]
    }]
}
```

### Course Update Task (CourseUpdateTask)

Course update tasks will update a course from source and perform the standard update protocol
(syncing with the LMS, build assignment images, etc).
This task has no additional configuration.

Basic Example:
```json
{
    ... the rest of a course object ...
    "course-update": [{
        "when": [{
            "daily": "02:00"
        }]
    }]
}
```

### Report Task (ReportTask)

The report task sends an email to the target users summarizing the current submissions for each assignment.

| Name         | Type        | Required | Description |
|--------------|-------------|----------|-------------|
| `to`         | List[Email] | true     | A list of emails to send the report to. |
| `send-empty` | Boolean     | false    | If true, the report will still be sent even if no submissions have been made. |

Basic Example:
```json
{
    ... the rest of a course object ...
    "report": [{
        "when": [{
            "daily": "03:00"
        }],
        "to": [
            "alice@test.edulinq.org",
            "bob@test.edulinq.org"
        ]
    }]
}
```

### Scoring Upload Task (ScoringUploadTask)

A scoring upload task performs a full scoring
(including reexamining already scored/due assignments) and uploads to scores to the course's LMS.
This task will also recompute late information in the case that any factors have changed
(e.g., a late submission was removed, a due date changed, or a student was given additional late days).
This task has no additional configuration options.

Basic Example:
```json
{
    ... the rest of a course object ...
    "scoring-upload": [{
        "when": [{
            "daily": "04:00"
        }]
    }]
}
```

### Email Logs Task (EmailLogsTask)

The email logs task sends an email to the target users containing matching logs.

| Name         | Type        | Required | Description |
|--------------|-------------|----------|-------------|
| `to`         | List[Email] | true     | A list of emails to send the results to. |
| `send-empty` | Boolean     | false    | If true, the email will still be sent even if no query results were found. |
| `query`      | LogQuery    | true     | The log query to execute. |

Basic Example:
```json
{
    ... the rest of a course object ...
    "email-logs": [{
        "when": [{
            "daily": "05:00"
        }],
        "to": [
            "alice@test.edulinq.org",
            "bob@test.edulinq.org"
        ],
        "query": {
            "past": "24h",
            "level": "ERROR"
        }
    }]
}
```

An example using multiple queries.
Email Alice every day with the day's ERROR logs,
and email Bob every hour with the hour's WARN (which include ERROR) logs.
```json
{
    ... the rest of a course object ...
    "email-logs": [
        {
            "when": [{
                "daily": "05:00"
            }],
            "to": [
                "alice@test.edulinq.org"
            ],
            "query": {
                "past": "24h",
                "level": "ERROR"
            }
        },
        {
            "when": [{
                "every": {
                    "hours": 1
                }
            }],
            "to": [
                "bob@test.edulinq.org"
            ],
            "query": {
                "past": "1h",
                "level": "WARN"
            }
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

## LMS Adapter (LMSAdapter)

TODO

## Late Policy (LatePolicy)

TODO

## Submission Limit (SubmissionLimit)

TODO

## File Specification (FileSpec)

TODO

## File Operation (FileOP)
