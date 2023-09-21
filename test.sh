#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function main() {
    if [[ $# -gt 1 ]]; then
        echo "USAGE: $0 [test regex]"
        exit 1
    fi

    trap exit SIGINT

    local testRegex=$1

    cd "${THIS_DIR}"

    local options=""
    if [[ ! -z "${testRegex}" ]] ; then
        options="-run ${testRegex}"
    fi

    go test -v -count=1 ${options} ./...
    if [[ ${?} -ne 0 ]] ; then
        echo "Found test issues."
        return 1
    else
        echo "No issues found."
        return 0
    fi
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
