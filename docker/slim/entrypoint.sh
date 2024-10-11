#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly CMD_DIR="${THIS_DIR}/cmd"

readonly DATA_DIR='/data'

function main(){
    if [[ $# -eq 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    local executable=$1
    shift 1

    if [ ! -d "${CMD_DIR}/${executable}" ]; then
        echo "Could not find command: '${executable}'"
        exit 1
    fi

    go run ${CMD_DIR}/${executable}/main.go -c "dirs.base=${DATA_DIR}" "$@"
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
