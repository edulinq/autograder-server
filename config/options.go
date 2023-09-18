package config

var (
    COURSES_ROOT = newOption("courses.rootdir", "courses", "The default places to look for courses.");

    DEBUG = newOption("debug", false, "Enable general debugging.");

    DOCKER_DISABLE = newOption("docker.disable", false, "Disable the use of docker (usually for testing).");

    NO_STORE = newOption("grader.nostore", false, "Do not store grading outout (submissions).");

    LOG_LEVEL = newOption("log.level", "INFO", "The default logging level.");

    TESTS_DIR = newOption("test.tests_dir", "_tests", "The base dir containing courses and assignments for testing.");

    WEB_PORT = newOption("web.port", 8080, "The port for the web interface to serve on.");
)
