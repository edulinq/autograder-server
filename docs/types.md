# Types

Descriptions of types (usually called "configs") are always in JSON.

TODO: ToC

## Semantic Types

These are types that have simple underlying types (int, string, etc),
but adds a layer of semantics on top of it.
These are usually used as part of larger types.

`identifier`
String
IDs must only have letters, digits, and single sequences of periods, underscores, and hyphens
IDs cannot start or end with periods, underscores, or hyphens

`timestamp`
Int64
MS since UNIX epoch (which is UTC).

Pointer
This is a modifier to any existing type.
A pointer can take either the base value or "null".
There is no need to do any refernginc (e.g., `&`) in your JSON, the autograder will handle that automatically.
We will denote it with a start preceding the type name.
For example, a \*int can take a normal int (e.g. `123`) or a `null` value.

## Course

TOOD

## Assignment

JSON

| Name               | Type               | Required | Default Value | Synced from LMS | Description |
|--------------------|--------------------|----------|---------------|-----------------|-------------|
| `id`               | identifier         | true     |               | false           | Identifier for an assignment. Must be unique within a server. |
| `name`             | string             | false    | `<id>`        | true            | Display name for an assignment. Defaults to the assignment's identifier. |
| `sort-id`          | string             | false    |               | false           | An optional ID to use when sorting assignments. If not provided, an assignment's id will be used when ordering is required. |
| `due-date`         | \*timestamp        | false    | `null`        | true            | The due data for an assignment. This can be synced from the course LMS. |
| `max-points`       | float              | false    |               | true            | The maximum number of points available for the assignment. Although not required when grading, some late policies need this. |
| `lms-id`           | string             | false    |               | true*           | The LMS idendifier for this assignment. May be synced with the LMS if the assignment's name matches. |
| `late-policy`      | \*late-policy      | false    | `null`        | false           | The late policy to use for this assignment. Overrides any late policy set on the course level. |
| `submission-limit` | \*submission-limit | false    | `null`        | false           | The submission limit to enforce for this assignment. Overrides any limits set on the course level. |
| `image`            | string             | true     |               | false           | The base Docker image to use for this assignment. |
| `pre-static-docker-commands`  | list[string]   | false | []        | false           | A list of Docker commands to run before static files are copied into the image. |
| `post-static-docker-commands` | list[string]   | false | []        | false           | A list of Docker commands to run after static files are copied into the image. |
| `invocation`                  | list[string]   | false | []        | false           | The command to run when the grading container is launched (Docker's `CMD`). If not set, the image's default `ENTRYPOINT`/`CMD` is used. |
| `static-files`                | list[filespec] | false | []        | false           | A list of files to copy into the image's `/work` directory. |
| `pre-static-files-ops`        | list[fileop]   | false | []        | false           | A list of file operations to run before static files are copied into the image. |
| `post-static-files-ops`       | list[fileop]   | false | []        | false           | A list of file operations to run after static files are copied into the image. |
| `post-submission-files-ops`   | list[fileop]   | false | []        | false           | A list of file operations to run after a student's code is available in the `/input` directory. |

LMS sync.
`lms-id` or name match.

TODO

Build stages
pre-static docker
pre-static file ops
static
post-static file ops
post-static docker
invoation
submission

## Late Policy

TODO

## Submission Limits

TODO

## File Specification (filespec)

TODO

## File Operation (fileop)
