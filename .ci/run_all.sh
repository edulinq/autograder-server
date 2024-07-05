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

    echo "Building"
    ./scripts/build.sh
    if [[ $? -ne 0 ]] ; then
        echo "Failed to build the autograder."
        exit 1
    fi

    echo -e "\nInstalling Python Dependencies"
    ./.ci/install_py_interface.sh
    if [[ $? -ne 0 ]] ; then
        echo "Failed to install the Python interface."
        exit 1
    fi

    local error_count=0

    echo -e "\nChecking Formatting"
    ./scripts/check_formatting.sh
    ((error_count += $?))

    echo -e "\nChecking Duplicate API Locators"
    ./scripts/check_duplicate_api_locators.sh
    ((error_count += $?))

    echo -e "\nRunning Tests"
    ./scripts/run_tests.sh
    ((error_count += $?))

    echo -e "\nRunning Remote Tests"
    ./.ci/run_remote_tests.sh
    ((error_count += $?))

    echo -e "\nVerifying Python Test Data"
    ./.ci/verify_py_test_data.sh
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
