#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly BIN_DIR="${THIS_DIR}/bin"

function main(){
    if [[ $# -eq 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    local executable=$1
    shift 1

    if [ ! -f "${BIN_DIR}/${executable}" ]; then
        echo "Could not find command: '${executable}'"
        exit 1
    fi

    "${BIN_DIR}/${executable}" "$@"
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
