package config

var (
    COURSES_ROOT = newOption("courses.rootdir", "courses", "The default places to look for courses.");

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

    SECRETS_PATH = newOption("secrets.path", ".secrets.json", "Path to an optional secrets config file.");

    TESTS_DIR = newOption("test.tests_dir", "_tests", "The base dir containing courses and assignments for testing.");

    WEB_PORT = newOption("web.port", 8080, "The port for the web interface to serve on.");
)
