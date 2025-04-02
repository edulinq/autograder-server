#!/bin/bash

# Ensure that all test submissions run correctly.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function check_grader() {
    local assignment_dir="${1}"

    local assignment_path="${assignment_dir}/assignment.json"
    local submissions_dir="${assignment_dir}/sample-submissions"

    python3 -m autograder.cli.testing.test-submissions -a "${assignment_path}" -s "${submissions_dir}"
    local error_count=$?

    if [[ ${error_count} -gt 0 ]] ; then
        echo "ERROR: Grader test failed for assignment: '${assignment_dir}'."
    fi

    return ${error_count}
}

function check_assignments() {
    local error_count=0

    # Assignment dirs are identified by an assignment.json file.
    for assignment_path in "${THIS_DIR}/"*"/assignment.json" ; do
        assignment_dir=$(dirname "${assignment_path}")
        echo "Checking assignment: '${assignment_dir}'."

        check_grader "${assignment_dir}"
        ((error_count += $?))
    done

    return ${error_count}
}

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    check_assignments
    local error_count=$?

    if [[ ${error_count} -gt 0 ]] ; then
        echo "Found ${error_count} issues."
    else
        echo "No issues found!"
    fi

    return ${error_count}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
