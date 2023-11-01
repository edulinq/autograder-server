package config

import (
    "os"
    "path/filepath"
)

var (
    NO_AUTH = mustNewOption("web.noauth", false, "Disbale authentication on the API.");

    COURSES_ROOT = mustNewOption("courses.rootdir", "_courses", "The default places to look for courses.");

    WORK_DIR = mustNewOption("dirs.work", filepath.Join(os.TempDir(), "_autograder"), "The root dir roe autograder output and artifacts.");

    DEBUG = mustNewOption("debug", false, "Enable general debugging.");

    DOCKER_DISABLE = mustNewOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    EMAIL_FROM = mustNewOption("email.from", "", "From address for emails sent from the autograder.");
    EMAIL_HOST = mustNewOption("email.host", "", "SMTP host for emails sent from the autograder.");
    EMAIL_PASS = mustNewOption("email.pass", "", "SMTP password for emails sent from the autograder.");
    EMAIL_PORT = mustNewOption("email.port", "", "SMTP port for emails sent from the autograder.");
    EMAIL_USER = mustNewOption("email.user", "", "SMTP username for emails sent from the autograder.");

    NO_STORE = mustNewOption("grader.nostore", false, "Do not store grading outout (submissions).");

    LOG_LEVEL = mustNewOption("log.level", "INFO", "The default logging level.");
    LOG_PRETTY = mustNewOption("log.pretty", true, "Make the logging human-readable, but less efficient.");

    LOCAL_CONFIG_PATH = mustNewOption("config.local.path", "config.json", "Path to an optional local config file.");
    SECRETS_CONFIG_PATH = mustNewOption("config.secrets.path", ".secrets.json", "Path to an optional secrets config file.");

    BACKUP_DIR = mustNewOption("server.backup.dir", "_backup", "Path to where backups are made by default.");

    NO_TASKS = mustNewOption("tasks.disable", false, "Disable all scheduled tasks.");

    TESTING_MODE = mustNewOption("testing", false, "Assume tests are being run, which may alter some operations.");

    WEB_PORT = mustNewOption("web.port", 8080, "The port for the web interface to serve on.");
)
