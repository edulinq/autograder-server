#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly BIN_DIR="${ROOT_DIR}/bin"
readonly CMD_DIR="${ROOT_DIR}/cmd"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    set -e
    trap exit SIGINT

    cd "${ROOT_DIR}"

    mkdir -p "${BIN_DIR}"

    for main_path in "${CMD_DIR}/"*"/main.go" ; do
        local bin_name=$(dirname "${main_path}" | xargs basename)

        go build -o "${BIN_DIR}/${bin_name}" "${main_path}"
    done
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
