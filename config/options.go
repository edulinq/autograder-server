package config

import (
    "os"
    "path/filepath"
)

var (
    NO_AUTH = newOption("web.noauth", false, "Disbale authentication on the API.");

    COURSES_ROOT = newOption("courses.rootdir", "_courses", "The default places to look for courses.");

    WORK_DIR = newOption("dirs.work", filepath.Join(os.TempDir(), "_autograder"), "The root dir roe autograder output and artifacts.");
    CACHE_DIR = newOption("dirs.cache", filepath.Join(WORK_DIR.GetString(), "cache"), "A place for the autograder to store information to be cached between restarts.");

    DEBUG = newOption("debug", false, "Enable general debugging.");

    DOCKER_DISABLE = newOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    EMAIL_FROM = newOption("email.from", "", "From address for emails sent from the autograder.");
    EMAIL_HOST = newOption("email.host", "", "SMTP host for emails sent from the autograder.");
    EMAIL_PASS = newOption("email.pass", "", "SMTP password for emails sent from the autograder.");
    EMAIL_PORT = newOption("email.port", "", "SMTP port for emails sent from the autograder.");
    EMAIL_USER = newOption("email.user", "", "SMTP username for emails sent from the autograder.");

    NO_STORE = newOption("grader.nostore", false, "Do not store grading outout (submissions).");

    LOG_LEVEL = newOption("log.level", "INFO", "The default logging level.");
    LOG_PRETTY = newOption("log.pretty", true, "Make the logging human-readable, but less efficient.");

    LOCAL_CONFIG_PATH = newOption("config.local.path", "config.json", "Path to an optional local config file.");
    SECRETS_CONFIG_PATH = newOption("config.secrets.path", ".secrets.json", "Path to an optional secrets config file.");

    BACKUP_DIR = newOption("server.backup.dir", "_backup", "Path to where backups are made by default.");

    NO_TASKS = newOption("tasks.disable", false, "Disable all scheduled tasks.");

    WEB_PORT = newOption("web.port", 8080, "The port for the web interface to serve on.");
)
