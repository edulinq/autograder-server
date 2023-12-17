#!/bin/bash

# Verify the Python interface's test data against the server.
# The python interface should have already been installed via '.ci/install-py-interface.sh'.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BASE_DIR="${THIS_DIR}/.."

readonly PORT='12345'
readonly TESTS_DIR="${BASE_DIR}/_tests"

readonly TEMP_DIR='/tmp/__autograder__/autograder-py'

readonly SERVER_URL="http://127.0.0.1:${PORT}"

function verify_test_data() {
    echo "Verifying test data."

    cd "${TEMP_DIR}"

    ./.ci/verify_test_api_requests.py --server "${SERVER_URL}"
    return $?
}

function run() {
    cd "${BASE_DIR}"
    echo "Building project."

    ./build.sh
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to build server."
        return 100
    fi

    cd "${BASE_DIR}"

    # Start the server.
    ./bin/server \
        -c web.port="${PORT}" \
        -c log.level=DEBUG \
        --unit-testing &
    local server_pid="$!"
    sleep 1

    local error_count=0

    verify_test_data
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

    run
    ((error_count += $?))

    if [[ ${error_count} -gt 0 ]] ; then
        echo "Found ${error_count} issues."
    else
        echo "No issues found."
    fi

    return ${error_count}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
