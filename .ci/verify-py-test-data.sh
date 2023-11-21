#!/bin/bash

# Verify the Python interface's test data against the server.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BASE_DIR="${THIS_DIR}/.."

readonly PORT='12345'
readonly TESTS_DIR="${BASE_DIR}/_tests"

readonly TEMP_DIR='/tmp/__autograder__/verify_test_data'

readonly REPO_URL='https://github.com/eriq-augustine/autograder-py.git'

readonly SERVER_URL="http://127.0.0.1:${PORT}"

function verify_test_data() {
    echo "Verifying test data."

    cd "${TEMP_DIR}"

    ./.ci/verify_test_api_requests.py --server "${SERVER_URL}"
    return $?
}

function fetch_repo() {
    echo "Fetching Python interface repo."

    if [[ -d "${TEMP_DIR}" ]] ; then
        echo "Found existing repo '${TEMP_DIR}', skipping clone/checkout."
        return 0
    fi

    mkdir -p "${TEMP_DIR}"

    git clone "${REPO_URL}" "${TEMP_DIR}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to clone repo."
        return 1
    fi

    local branch=$(git branch --show-current)

    cd "${TEMP_DIR}"

    git checkout "${branch}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to checkout target branch ('${branch}')."
        return 2
    fi

    return 0
}

function run() {
    cd "${BASE_DIR}"
    echo "Building project."

    ./build.sh
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to build server."
        return 100
    fi

    fetch_repo
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to fetch the Python interface repo."
        return 110
    fi

    cd "${BASE_DIR}"

    # Start the server.
    ./bin/server \
        -c web.port="${PORT}" \
        -c web.noauth=true \
        -c courses.rootdir="${TESTS_DIR}" \
        -c grader.nostore=true \
        -c log.level=DEBUG &
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
