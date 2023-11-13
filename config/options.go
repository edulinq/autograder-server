package config

import (
    "os"
    "path/filepath"
)

var (
    COURSES_ROOT = MustNewStringOption("courses.rootdir", "_courses", "The default places to look for courses.");

    WORK_DIR = MustNewStringOption("dirs.work", filepath.Join(os.TempDir(), "_autograder"), "The root dir for autograder output and artifacts.");

    DEBUG = MustNewBoolOption("debug", false, "Enable general debugging.");

    DOCKER_DISABLE = MustNewBoolOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    EMAIL_FROM = MustNewStringOption("email.from", "", "From address for emails sent from the autograder.");
    EMAIL_HOST = MustNewStringOption("email.host", "", "SMTP host for emails sent from the autograder.");
    EMAIL_PASS = MustNewStringOption("email.pass", "", "SMTP password for emails sent from the autograder.");
    EMAIL_PORT = MustNewStringOption("email.port", "", "SMTP port for emails sent from the autograder.");
    EMAIL_USER = MustNewStringOption("email.user", "", "SMTP username for emails sent from the autograder.");

    NO_STORE = MustNewBoolOption("grader.nostore", false, "Do not store grading outout (submissions).");

    LOG_LEVEL = MustNewStringOption("log.level", "INFO", "The default logging level.");
    LOG_PRETTY = MustNewBoolOption("log.pretty", true, "Make the logging human-readable, but less efficient.");

    LOCAL_CONFIG_PATH = MustNewStringOption("config.local.path", "config.json", "Path to an optional local config file.");
    SECRETS_CONFIG_PATH = MustNewStringOption("config.secrets.path", ".secrets.json", "Path to an optional secrets config file.");

    BACKUP_DIR = MustNewStringOption("server.backup.dir", "_backup", "Path to where backups are made by default.");

    NO_TASKS = MustNewBoolOption("tasks.disable", false, "Disable all scheduled tasks.");

    TESTING_MODE = MustNewBoolOption("testing", false, "Assume tests are being run, which may alter some operations.");

    WEB_PORT = MustNewIntOption("web.port", 8080, "The port for the web interface to serve on.");
    NO_AUTH = MustNewBoolOption("api.noauth", false, "Disbale authentication on the API.");

    DB_TYPE = MustNewStringOption("db.type", "disk", "The type of database to use.");
    DB_DISK_DIR = MustNewStringOption("db.disk.dir", "disk-database", "The dir (absolute or relative to WORK_DIR) for the disk database.");
    DB_SQLITE_PATH = MustNewStringOption("db.sqlite.path", "autograder.sqlite3", "The path (absolute or relative to WORK_DIR) for the SQLite3 database.");
    DB_PG_URI = MustNewStringOption("db.pg.uri", "", "Connection string to connect to a Postgres Databse. Empty if not using Postgres.");
)
