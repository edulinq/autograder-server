#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ASSIGNMENT_CONFIG="${THIS_DIR}/assignment.json"

function main() {
    if [[ $# -eq 0 ]]; then
        echo "USAGE: $0 <submission dir>"
        exit 1
    fi

    set -e
    trap exit SIGINT

    local submission_dir=$1
    shift 1

    python3 -m autograder.cli.grading.grade -a "${ASSIGNMENT_CONFIG}" -s "${submission_dir}" $@
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
