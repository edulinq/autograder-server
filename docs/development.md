# Development Notes

## Memory Persistence Mentality

In general, we want to avoid keeping things in memory for more than they are absolutely necessary.
For most objects, this means that we only want them in memory for the duration of an API request.
(Since we do not need to heavily optimize for throughput and our objects tend to be simple,
this should not be a performance concern).
This mentality will allow us to "live load" things like courses, assignment, and users.
Changes to them should be reflected in the next use (usually an API call).
This means that most actions will begin with a call to the DB so the required object can be loaded into memory.

## "Must" and "Should" Functions

In Go, there is a convention of "Must" functions (e.g. `regexp.MustCompile()`).
These methods are supposed to either complete their task, or panic.

In the autograder, "Must" functions should only be used in testing code,
in a program's main file (excluding the server),
and code that is verified once run a single time (in a test).
The last situation is where arguments that can cause the possible error are fixed
and the code just needs to be run once to verify that it will not generate an error.
The prime example of this is `regexp.MustCompile()`, where the regexp is a string literal which cannot change.
Once the regexp is verified to compile once (in a test), it should compile every time.

In addition to "Must" functions, this project also uses "Should" functions.
A "Should" function is for situations where a function could encounter an error in very rare circumstances,
but may still produce a workable result.
In the case of an error, a "Should" method will log the error and return a reasonable result.
An example is `util.ShouldAbs()`, which computes an absolute path from a given path.
If the current working directory cannot be fetched, then `filepath.Abs()` will  return an error.
Since this is very unlikely and most times an abs path is not strictly necessary,
`util.ShouldAbs()` can just log an error and return the original path.

"Should" functions can be used anywhere,
but care should be taken not to overuse them when a fallback value is not sufficient.
"Must" variants should be favored in testing code, since they will more obviously fail tests.

## Internal Notes

### Passwords/Tokens

Cleartext is never stored in anything other than memory (and event then it will be temporary).
When we are working with a "password", it should always be a [SHA-256](https://en.wikipedia.org/wiki/SHA-2) hash of the cleartext.
Any variables/arguments actually using cleartext should make that apparent with the variable name (i.e. the name should have "cleartext" in it).

## Test Data

To test our code, we supply test data, which can be found in the `testdata` directory. There are test courses and test users, as described below.

### Test Courses

Test courses allow testing of course specific code. The default test course is "course101" which contains a user for every course role.
course101 includes a sample assignment, "HW0", and 3 test submissions made by "course-student@test.edulinq.org".
There is another test course that has assignments using different languages: "course-languages".

### Test Users

Test users allow testing of actions made by the various role users can hold on the server.
By default, the name and password of test users are their email without "@test.edulinq.org".

Users prefixed with "server-" (i.e. "server-owner@test.edulinq.org") are used to test the corresponding server roles.
These users are not enrolled in any test courses by default.

Users prefixed with "course-" (i.e. "course-owner@test.edulinq.org") are used to test the corresponding course roles.
All of these users are given the standard server role, which is server user.
To test courses actions, these users are enrolled in various test courses.

## Locatable Errors

All locators are negative numbers and only exist within `internal`.
Each top-level package within `internal` can be allocated blocks of 1000 locators.
See the table below for blocks that are already allocated locators.
After being allocated locators, the package can allocate sub-blocks of the locators to subpackages.
Use `scripts/get_max_locators.sh` to determine the next locator to use within a certain package.
If a package you are working on requires locatable errors, the package gets the next chunk of 1000 locators.
Update this document to include the new top-level allocation.

### Top-Level Package Allocations

|Package    |Upper Bound |Lower Bound |
|-----------|------------|------------|
|api        |-001        |-999        |
|procedures |-1000       |-1999       |
|lms        |-2000       |-2999       |

### API Errors

All API errors are locatable errors.
We allocate 3 digit negative numbers (-001 to -999) as the locators for all API errors.
Within this range, each package in the API package is given a range of 100 locators.

## API Notes

### Passwords/Tokens

No password or token should be sent to the server as cleartext.
They should always be hashed using the [SHA-256](https://en.wikipedia.org/wiki/SHA-2) cryptographic hash function.
On request, cleartext passwords and tokens may be set from the server to the user (or via email),
but never in the other direction.

### Role Escalation

API requests that are at least course user context must be called on a user that is enrolled in the course.
However, users with at least server role admin are allowed to call any API.

To achieve this, we convert these high level server users to course owners during the validation of the request.
A server admin (or above) that is not enrolled in the course will have a nil LMSID.
If they are enrolled in the course, their role will be set to course owner, regardless of their existing course role.

## Cross-Platform Scripting

We are writing bash scripts with the intention of running them in a POSIX environment.
However, differences between operating systems, particularly between Linux and BSD,
can still affect a script's behavior.
This section highlights key considerations to ensure your scripts run consistently across platforms.

### Regular Expressions

Differences between how Linux and BSD handle different standards of regular expressions can cause inconsistent outcomes when writing scripts.
Here are three common regex standards you may encounter, along with key considerations for each:

- **BRE (Basic Regular Expressions)**
  - **Usage**: Default in many tools.
  - **Consideration**: While BREs cover most common use cases, their default behavior can vary between tools and operating systems.
  - **[More Info](https://en.wikipedia.org/wiki/Regular_expression#IEEE_POSIX_Standard)**

- **ERE (Extended Regular Expressions)**
  - **Usage**: Often enabled with the `-E` flag (like in `grep` and `sed`).
  - **Consideration**: EREs make it easier to write and understand complex regular expressions.
  - **[More Info](https://en.wikipedia.org/wiki/Regular_expression#IEEE_POSIX_Standard)**

- **PCRE (Perl-Compatible Regular Expressions)**
  - **Usage**: Enabled with the `-P` flag in `grep`.
  - **Consideration**: PCREs have inconsistent behavior between Linux and BSD. Avoid using for cross-platform scripts.
  - **[More Info](https://en.wikipedia.org/wiki/Perl_Compatible_Regular_Expressions)**

### Tool-Specific Guidelines

#### `grep`
- **Avoid**: The `-P` flag (PCRE) for cross-platform scripts.
- **Alternative**: Use the `-E` flag (ERE) to achieve similar functionality with more consistent behavior.

#### `sed`
- **Use**: The `-E` flag to enable ERE.
- **Reason**: BRE behavior is inconsistent across different operating systems.

## Shared Working Directory for Server Interactions

A server instance is defined by its working directory,
so anything that wants to interact with a server (including a cmd) needs to make sure they share the same working directory.
This can be done by setting the base directory in the command line whenever interacting with the server (-c dirs.base).
For example:
```
go run cmd/peek-submission/main.go course-student@test.edulinq.org course101 hw0 -c dirs.base="/path/to/working/directory"
```

