package config

var (
    // Base
    NAME = MustNewStringOption("instance.name", "autograder",
        "A name to identify this autograder instance." +
        " Should only contain alphanumerics and underscores.");

    // Directories
    BASE_DIR = MustNewStringOption("dirs.base", GetDefaultBaseDir(),
        "The base dir for autograder to store data." +
        " SHOULD NOT be set in config files (to prevent cycles), only on the command-line.");

    // Debug / Testing
    DEBUG = MustNewBoolOption("debug", false, "Enable general debugging.");
    NO_STORE = MustNewBoolOption("grader.nostore", false, "Do not store grading output (submissions).");
    TESTING_MODE = MustNewBoolOption("testing", false, "Assume tests are being run, which may alter some operations.");
    NO_AUTH = MustNewBoolOption("api.noauth", false, "Disable authentication on the API.");
    STORE_HTTP = MustNewStringOption("http.store", "", "Store HTTP requests made by the server to the specified directory.");

    // Logging
    LOG_TEXT_LEVEL = MustNewStringOption("log.text.level", "INFO", "The default logging level for the text (stderr) logger.");
    LOG_BACKEND_LEVEL = MustNewStringOption("log.backend.level", "INFO", "The default logging level for the backend (database) logger.");

    // Email
    EMAIL_FROM = MustNewStringOption("email.from", "", "From address for emails sent from the autograder.");
    EMAIL_HOST = MustNewStringOption("email.host", "", "SMTP host for emails sent from the autograder.");
    EMAIL_PASS = MustNewStringOption("email.pass", "", "SMTP password for emails sent from the autograder.");
    EMAIL_PORT = MustNewStringOption("email.port", "", "SMTP port for emails sent from the autograder.");
    EMAIL_USER = MustNewStringOption("email.user", "", "SMTP username for emails sent from the autograder.");

    // Docker
    DOCKER_DISABLE = MustNewBoolOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    // Tasks
    NO_TASKS = MustNewBoolOption("tasks.disable", false, "Disable all scheduled tasks.");
    TASK_MIN_REST_SECS = MustNewIntOption("tasks.minrest", 5 * 60,
            "The minimum time (in seconds) between invocations of the same task." +
            " A task instance that tries to run too quickly will be skipped.");
    TASK_BACKUP_DIR = MustNewStringOption("tasks.backup.dir", "", "Path to where backups are made. Defaults to inside BASE_DIR.");

    // Server
    WEB_PORT = MustNewIntOption("web.port", 8080, "The port for the web interface to serve on.");
    WEB_MAX_FILE_SIZE_KB = MustNewIntOption("web.maxsizekb", 2 * 1024, "The maximum allowed file size (in KB) submitted via POST request. The default is 2048 KB (2 MB).");

    // Database
    DB_TYPE = MustNewStringOption("db.type", "disk", "The type of database to use.");
    DB_PG_URI = MustNewStringOption("db.pg.uri", "", "Connection string to connect to a Postgres Databse. Empty if not using Postgres.");

    STALELOCK_DURATION_SECS = MustNewIntOption("lockmanager.staleduration", 2 * 60 * 60, "Number of seconds a lock can be unused before getting removed.");
)
