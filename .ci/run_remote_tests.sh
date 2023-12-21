#!/bin/bash

# Runs submissions tests by running a local instance of the server submitting agasint that.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BASE_DIR="${THIS_DIR}/.."

readonly PORT='12345'
readonly TESTS_DIR="${BASE_DIR}/_tests"

readonly COURSE_CONFIG_FILENAME='course.json'
readonly ASSIGNMENT_CONFIG_FILENAME='assignment.json'
readonly TEST_SUBMISSIONS_DIRNAME='test-submissions'

readonly DEFAULT_SERVER="http://127.0.0.1:${PORT}"

# These do not actually matter, the server is run in noauth mode.
readonly TEST_USER='admin@test.com'
readonly TEST_PASS='admin'

function run_submissions() {
    local error_count=0
    local run_count=0

    for assignment_config_path in $(find "${TESTS_DIR}" -type f -name "${ASSIGNMENT_CONFIG_FILENAME}") ; do
        local course_config_path="$(dirname "${assignment_config_path}")/../${COURSE_CONFIG_FILENAME}"
        if [[ ! -f "${course_config_path}" ]] ; then
            echo "ERROR: Cannot find course config for assignment ('${assignment_config_path}')."
            ((error_count += 1))
            continue
        fi

        local submission_dir="$(dirname "${assignment_config_path}")/${TEST_SUBMISSIONS_DIRNAME}"
        if [[ ! -d "${submission_dir}" ]] ; then
            echo "Assignment ('${assignment_config_path}') does not have any submissions."
            continue
        fi

        local assignment_id=$(grep '"id"' "${assignment_config_path}" | sed 's/^\s*"id"\s*:\s*"\([^"]\+\)\s*",\s*$/\1/')
        if [[ -z "${assignment_id}" ]] ; then
            echo "ERROR: Could not find assignment ID for '${assignment_config_path}'."
            ((error_count += 1))
            continue
        fi

        local course_id=$(grep '"id"' "${course_config_path}" | sed 's/^\s*"id"\s*:\s*"\([^"]\+\)\s*",\s*$/\1/')
        if [[ -z "${course_id}" ]] ; then
            echo "ERROR: Could not find course ID for '${course_config_path}'."
            ((error_count += 1))
            continue
        fi

        echo "Testing assignment '${assignment_config_path}' on submissions '${submission_dir}'."

        python3 -m autograder.cli.testing.test-remote-submissions \
            --server "${DEFAULT_SERVER}" \
            --user "${TEST_USER}" \
            --pass "${TEST_PASS}" \
            --course "${course_id}" \
            --assignment "${assignment_id}" \
            "${submission_dir}"

        ((error_count += $?))
        ((run_count += 1))

        echo "---------------"
    done

    if [[ ${run_count} -eq 0 ]] ; then
        echo "ERROR: Cound not find any test submissions."
        ((error_count += 1))
    fi

    return ${error_count}
}

function run_sever_submissions() {
    cd "${BASE_DIR}"

    ./build.sh
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to build server."
        return 100
    fi

    # Start the server.
    ./bin/server \
        -c web.port="${PORT}" \
        -c log.level=DEBUG \
        --unit-testing &
    local server_pid="$!"
    sleep 1

    local error_count=0

    run_submissions
    ((error_count += $?))

    # Stop the server.
    kill "${server_pid}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Server did not terminate cleanly."
        ((error_count++))
    fi

    sleep 1

    return ${error_count}
}

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    run_sever_submissions
    ((error_count += $?))

    if [[ ${error_count} -gt 0 ]] ; then
        echo "Found ${error_count} issues."
    else
        echo "No issues found."
    fi

    return ${error_count}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
