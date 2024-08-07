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

We are writing bash scripts with the intention of running them in a POSIX environment. However, differences between operating systems, particularly between Linux and MacOS, can still affect a script's behavior. This section highlights key considerations to ensure your scripts run consistently across platforms.

### Grep

The `-P` flag, which enables perl-compatible regular expressions, does not work consistently across operating systems. More specifically, while it may work as intended on Linux, it behaves differently on Mac. This can lead to unexpected results when run on MacOS. To avoid inconsistency across operating systems, avoid using the `-P` flag. Sometimes, using the `-E` flag, which enables extended regular expressions, may produce the same output that `-P` was going for. 

### Sed and Regular Expression

When using regular expressions in shell commands with tools like `sed`, there are differences between how Linux and MacOS handle them due to differences between basic regular expressions and extended regular expressions.

* **BRE (Basic Regular Expressions)**: Requires escaping certain metacharacters such as `+` to be able to interpret them as special characters. 
* **ERE (Extended Regular Expressions)**: Does not require escaping these metacharacters since they are treated as special characters by default.

However, something like `\+` that should be interpreted as "one or more" in a BRE does not work consistently on Mac. To ensure consistency across all operating systems, the `-E` flag should be used when using special characters, which would make `+` work as "one or more."