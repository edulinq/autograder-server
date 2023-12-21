#!/bin/bash

# Compile the project and run the grader.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DEFAULT_OUTPUT_PATH="${THIS_DIR}/../output/result.json"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        return 1
    fi

    trap exit SIGINT

    cd "${THIS_DIR}"

    # Allow the grader to run locally by changing the output location
    # if not in the docker image.
    local outputPath="${DEFAULT_OUTPUT_PATH}"
    if [[ ! -d $(dirname "${outputPath}") ]] ; then
        outputPath="$(basename "${outputPath}")"
    fi

    # Compile.
    javac -cp .:json-20231013.jar *.java
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to compile assignment."
        return 2
    fi

    # Run grader.
    java -cp .:json-20231013.jar Grader > "${outputPath}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to run grader."
        return 3
    fi

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
