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

## Cross-Platform Scripting

We are writing bash scripts with the intention of running them in a POSIX environment.
However, differences between operating systems, particularly between Linux and MacOS,
can still affect a script's behavior. 
This section highlights key considerations to ensure your scripts run consistently across platforms.

### Regular Expressions
Differences between how Linux and MacOS handle regular expressions can cause inconsistent outcomes when writing scripts.
Be aware of the following:

- **BRE (Basic Regular Expressions)**
  - **Usage**: Default in many tools.
  - **Consideration**: Limited feature set.
  - **[More Info](https://en.wikibooks.org/wiki/Regular_Expressions/Basic_Regular_Expressions)**

- **ERE (Extended Regular Expressions)**
  - **Usage**: Enabled with `-E` flag in tools like `grep` and `sed`.
  - **Consideration**: More expressive than BRE.
  - **[More Info](https://en.wikibooks.org/wiki/Regular_Expressions/POSIX-Extended_Regular_Expressions)**

- **PCRE (Perl-Compatible Regular Expressions)**
  - **Usage**: Enabled with `-P` flag in `grep`.
  - **Consideration**: Inconsistent behavior between Linux and MacOS. Avoid for cross-platform scripts.
  - **[More Info](https://en.wikibooks.org/wiki/Regular_Expressions/Perl-Compatible_Regular_Expressions)**

### Tool-Specific Guidelines

#### `grep`
- **Avoid**: `-P` flag (PCRE) for cross-platform scripts.
- **Alternative**: Use `-E` flag (ERE) to achieve similar functionality with more consistent behavior.

#### `sed`
- **Use**: `-E` flag to enable ERE.
- **Avoid**: BRE syntax that may not behave consistently across platforms (e.g., `\+` in BRE).