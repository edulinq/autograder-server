#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    local error_count=0

    echo "Running tests."
    go test -v -count=1 ./...
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
