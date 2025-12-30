#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly BIN_DIR="${ROOT_DIR}/bin"
readonly CMD_DIR="${ROOT_DIR}/cmd"

function main() {
    if [[ $# -gt 1 ]]; then
        echo "USAGE: $0 [bin name]"
        exit 1
    fi

    set -e
    trap exit SIGINT

    local target_name="${1}"

    cd "${ROOT_DIR}"

    mkdir -p "${BIN_DIR}"

    for main_path in "${CMD_DIR}/"*"/main.go" ; do
        local bin_name=$(dirname "${main_path}" | xargs basename)

        # If a bin name was given, skip other binaries.
        if [ ! -z "${target_name}" ] && [ "${target_name}" != "${bin_name}" ] ; then
            continue
        fi

        echo "Building ${bin_name}"
        CGO_ENABLED=0 go build -ldflags="-s -w" -o "${BIN_DIR}/${bin_name}" "${main_path}"
    done

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
