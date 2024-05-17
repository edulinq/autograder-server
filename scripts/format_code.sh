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

    echo "Formatting files ..."

    go fmt ./...
    local result=$?

    if [[ ${result} -ne 0 ]] ; then
        echo "Found issues while formatting files."
    fi

    return ${result}
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
