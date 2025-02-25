#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly OUT_DIR="${ROOT_DIR}/testdata/certs/ssl"

readonly KEY_PATH="${OUT_DIR}/test-ssl.key"
readonly CERT_PATH="${OUT_DIR}/test-ssl.crt"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT
    set -e

    cd "${ROOT_DIR}"

    mkdir -p "${OUT_DIR}"

    openssl genrsa -out "${KEY_PATH}" 2048

    openssl req -new \
        -x509 -sha256 \
        -key "${KEY_PATH}" \
        -out "${CERT_PATH}" \
        -days 3650 \
        -subj "/C=US/ST=California/L=Santa Cruz/O=EduLinq/OU=Autograder/CN=localhost"

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
