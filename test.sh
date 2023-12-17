#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${THIS_DIR}"

    local error_count=0

    # Run non-task test first.
    go test -v -count=1 -skip TestTask ./...
    if [[ ${?} -ne 0 ]] ; then
        ((error_count += 1))
    fi

    # Now run task tests.
    # The task tests are sensitive to CPU load and scheduling,
    # so should be run alone.
    go test -v -count=1 ./task -run TestTask
    if [[ ${?} -ne 0 ]] ; then
        ((error_count += 1))
    fi

    if [[ ${error_count} -gt 0 ]] ; then
        echo "Found test issues."
    else
        echo "No issues found."
    fi

    return ${error_count}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
