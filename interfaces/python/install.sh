#!/bin/bash

# Install via pip.
# The prerequisite package step clears the dist directory, so there should only be one entry.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    set -e
    trap exit SIGINT

    cd "${THIS_DIR}"

    rm -rf dist
    python3 setup.py sdist
    pip3 install dist/autograder-* --upgrade --force-reinstall
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
