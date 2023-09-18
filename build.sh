#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BIN_DIR="${THIS_DIR}/bin"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    set -e
    trap exit SIGINT

    mkdir -p "${BIN_DIR}"

    for main_path in "${THIS_DIR}/cmd/"*"/main.go" ; do
        local bin_name=$(dirname "${main_path}" | xargs basename)

        go build -o "${BIN_DIR}/${bin_name}" "${main_path}"
    done
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
