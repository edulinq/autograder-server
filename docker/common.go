package docker

import (
    "fmt"
    "strings"
)

const DOCKER_CONFIG_FILENAME = "config.json"

const DOCKER_BASE_DIR = "/autograder"
const DOCKER_INPUT_DIR = DOCKER_BASE_DIR + "/input"
const DOCKER_OUTPUT_DIR = DOCKER_BASE_DIR + "/output"
const DOCKER_WORK_DIR = DOCKER_BASE_DIR + "/work"
const DOCKER_CONFIG_PATH = DOCKER_BASE_DIR + "/" + DOCKER_CONFIG_FILENAME

func DockerfilePathQuote(path string) string {
    return fmt.Sprintf("\"%s\"", strings.ReplaceAll(path, "\"", "\\\""));
}
