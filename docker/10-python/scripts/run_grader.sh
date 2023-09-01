#!/bin/bash

# Check for assignment config files and run all tests in their directories.

readonly BASE_DIR="/autograder"

readonly INPUT_DIR="${BASE_DIR}/input"
readonly OUTPUT_DIR="${BASE_DIR}/output"
readonly WORK_DIR="${BASE_DIR}/work"

readonly CONFIG_PATH="${BASE_DIR}/config.json"
readonly GRADER_PATH="${WORK_DIR}/grader.py"
readonly OUTPUT_PATH="${OUTPUT_DIR}/result.json"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd ${BASE_DIR}

    python3 -m autograder.cli.pre-docker-grading
    if [[ $? -ne 0 ]] ; then
        echo "Failed to run docker pre-grading prep."
        exit 2
    fi

    python3 -m autograder.cli.grade-submission \
        --grader "${GRADER_PATH}" \
        --inputdir "${INPUT_DIR}" \
        --outputdir "${OUTPUT_DIR}" \
        --workdir "${WORK_DIR}" \
        --outpath "${OUTPUT_PATH}"
    local returnValue=$?

    if [[ ${returnValue} -ne 0 ]] ; then
        echo "Failed to run Python grader."
        exit ${returnValue}
    fi

    exit 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
