package config

var (
    LOG_LEVEL = newOption("log.level", "INFO", "The default logging level.");
    WEB_PORT = newOption("web.port", 8080, "The port for the web interface to serve on.");
    COURSES_ROOT = newOption("courses.rootdir", "courses", "The default places to look for courses.");

    DOCKER_DISABLE = newOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    TESTS_DIR = newOption("test.tests_dir", "../../autograder-py-tests", "The base dir containing courses and assignments for testing.");
)
