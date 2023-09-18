#!/bin/bash

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    set -e
    trap exit SIGINT

    go test -v -count=1 ./...
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
