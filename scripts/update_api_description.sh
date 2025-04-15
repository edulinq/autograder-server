#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly OUT_DIR="${ROOT_DIR}/resources"
readonly OUT_FILE="${OUT_DIR}/api.json"

function main() {
    if [[ $# -ne 0 ]] ; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    mkdir -p "${OUT_DIR}"

    go run cmd/describe-api-endpoints/main.go > "${OUT_FILE}"
    if [ $? -ne 0 ]; then
        echo "Failed to update ${OUT_FILE}."
        return 1
    fi

    echo "API description successfully updated in ${OUT_FILE}."
    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
