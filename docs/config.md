# Configuration

This project uses configuration options to set the behavior of its executables.
All executables that use autograder resources use the same configuration infrastructure
and can therefore be configured the same way and with the same options.
See the (Configuration Options section)[#configuration-options] for all available options.

Options can be set on the command-line using the `-c`/`--config` flag.
For example:
```
./bin/logs-example --config log.level=debug
```

Options can also be set using environmental variables by prefixing the option keys
with `AUTOGRADER__` and replacing any `.` with `__`.
For example option key `docker.disable` can be set by:
```
AUTOGRADER__DOCKER__DISABLE='true' ./scripts/run_tests.sh
```

## Directories

The primary directory the autograder will use for storing information is referred to as the "work directory",
and is set to `<base dir>/<instance name>`.
Most other paths are configured to be relative to the work directory.

The base directory is set through the `dirs.base` option,
and defaults to `$XDG_DATA_HOME`.

The instance name is a way to configuration a unique name for your autograder instance.
It can be set with the `instance.name` option,
and defaults to `autograder`.

## Loading Options

Configurations will be loaded in the following order (later options override earlier ones):
 0. The command-line options are checked for `BASE_DIR`.
 1. Load options from environmental variables.
 2. Options are loaded from `WORK_DIR/config` (config.json then secrets.json).
 3. Options are loaded from the current working directory (config.json then secrets.json).
 4. Options are loaded from any files specified with `--config-path` (ordered by appearance).
 5. Options are loaded from the command-line (`--config` / `-c`).

The base directory (`dirs.base`) can ONLY be set via the command-line or environmental variables.
This prevents cycles from the base directory changing and loading new options.

## Configuration Options

Below are all configuration options available for the autograder server.
You can also see all available options by looking in the [config/options.go](../internal/config/options.go) file,
or using the `cmd/list-options` executable.
```
go run cmd/list-options/main.go
```

| Key                          | Type    | Default Value  | Description |
|------------------------------|---------|----------------|-------------|
| `analysis.pairwise.poolsize` | Integer | 1               | The number of parallel workers per course when computing pairwise analysis. |
| `build.keep`                 | Boolean | false           | Keep artifacts/dirs used when building (not building the server itself, but things like assignment images). |
| `db.type`                    | String  | "disk"          | The type of database to use. |
| `db.pg.uri`                  | String  |                 | Connection string to connect to a Postgres Database. Empty if not using Postgres. |
| `dirs.base`                  | String  | [$XDG_DATA_HOME](https://specifications.freedesktop.org/basedir-spec/latest/) | The base dir for autograder to store data. SHOULD NOT be set in config files (to prevent cycles), only on the command-line. |
| `dirs.backup`                | String  | dirs.base       | Path to where backups are made. Defaults to inside BASE_DIR. |
| `docker.disable`             | Boolean | false           | Disable the use of docker (usually for testing). |
| `docker.output.maxsize`      | Integer | 4096 (4 MB)     | The maximum allowed size (in KB) for stdout and stderr combined. The default is 4096 KB (4 MB). |
| `email.from`                 | String  |                 | From address for emails sent from the autograder. |
| `email.host`                 | String  |                 | SMTP host for emails sent from the autograder. |
| `email.pass`                 | String  |                 | SMTP password for emails sent from the autograder. |
| `email.port`                 | String  |                 | SMTP port for emails sent from the autograder. |
| `email.user`                 | String  |                 | SMTP username for emails sent from the autograder. |
| `email.smtp.idle`            | Integer | 120000 (2 mins) | Consider an SMTP connection idle if no emails are sent for this number of milliseconds. |
| `email.smtp.minperiod`       | Integer | 250             | Allow for at least this amount of time (in milliseconds) between sending emails. |
| `grading.runtime.max`        | Integer | 300 (5 mins)    | The maximum number of seconds a grader can be running for. |
| `http.store`                 | String  |                 | Store HTTP requests made by the server to the specified directory. |
| `instance.name`              | String  | "autograder"    | A name to identify this autograder instance. Should only contain alphanumerics and underscores. |
| `lockmanager.staleduration`  | Integer | 7200 (2 hours)  | Number of seconds a lock can be unused before getting removed. |
| `log.text.level`             | String  | "INFO"          | The default logging level for the text (stderr) logger. |
| `log.backend.level`          | String  | "INFO"          | The default logging level for the backend (database) logger. |
| `tasks.disable`              | Boolean | false           | Disable all scheduled tasks. |
| `tasks.minrest`              | Integer | 300 (5 mins)    | The minimum time (in seconds) between invocations of the same task. A task instance that tries to run too quickly will be skipped. |
| `testing`                    | Boolean | false           | Assume tests are being run, which may alter some operations. |
| `testdata.load`              | Boolean | false           | Load test data when the database opens. |
| `web.port`                   | Integer | 8080            | The port for the web interface to serve on. |
| `web.maxsize`                | Integer | 2048 (2 MB)     | The maximum allowed file size (in KB) submitted via POST request. The default is 2048 KB (2 MB). |
| `web.static.root`            | String  |                 | The root directory to serve as part of the static portion of the API. Defaults to empty string, which indicates the embedded static directory. |
| `web.static.fallback`        | Boolean | false           | For any unmatched route (potential 404) that does not have an API prefix, try to match it in the static root before giving the final 404. |
