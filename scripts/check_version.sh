#!/bin/bash

# Compare the listed version against the most recent tagged version.

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly VERSION_PATH="${ROOT_DIR}/resources/VERSION.json"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    local taggedVersion=$(git tag -l | grep '^v' | sed 's/^v//' | sort -rV | head -n 1)
    if [[ ${taggedVersion} == "" ]] ; then
        echo "Could not get tagged version from git."
        return 2
    fi

    grep -q '"'"${taggedVersion}"'"' "${VERSION_PATH}"
    if [[ $? -ne 0 ]] ; then
        echo "Could not find tagged version '${taggedVersion}' inside of version file '${VERSION_PATH}'."
        return 3
    fi
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
