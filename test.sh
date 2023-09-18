#!/bin/bash

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    go test -v -count=1 ./...
    if [[ ${?} -ne 0 ]] ; then
        echo "Found test issues."
        return 1
    else
        echo "No issues found."
        return 0
    fi
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
