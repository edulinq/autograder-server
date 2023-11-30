#!/bin/bash

# Run all the tests.
# This is not used in CI so that each part can be run individually.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="${THIS_DIR}/.."

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    ./build.sh
    if [[ $? -ne 0 ]] ; then
        echo "Failed to build the autograder."
        exit 1
    fi

    ./.ci/install-py-interface.sh
    if [[ $? -ne 0 ]] ; then
        echo "Failed to install the Python interface."
        exit 1
    fi

    local error_count=0

    ./test.sh
    ((error_count += $?))

    ./.ci/run_remote_tests.sh
    ((error_count += $?))

    ./.ci/verify-py-test-data.sh
    ((error_count += $?))

    echo "================="

    if [[ ${error_count} -gt 0 ]] ; then
        echo "Found ${error_count} issues."
    else
        echo "No issues found."
    fi

    return ${error_count}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
