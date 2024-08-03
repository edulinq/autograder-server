#!/bin/bash

# Check for duplicate API locators.
# This does not work for locators that are on a different line from the locator creation function call.
# Specifically, we will look for 'Error("-' to indicate where locators are defined.

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly API_DIR="${ROOT_DIR}/internal/api"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    local duplicateLocators=$(grep -R 'Error("-' "${API_DIR}" | sed 's/^.*"\(-[0-9]\{3,\}\)".*$/\1/' | sort | uniq -c | grep -ZvE '^\s+1\s+' | sed 's/^\s*\([0-9]\+\)\s\+\(-[0-9]\{3,\}\)$/\2/')

    if [[ -z ${duplicateLocators} ]] ; then
        echo "No duplicate locators found."
        return 0
    fi

    echo "Found duplicate locators."

    for duplicateLocator in ${duplicateLocators} ; do
        echo "${duplicateLocator}:"
        grep -R --line-number "Error(\"${duplicateLocator}\"" "${API_DIR}" | sed 's/^/\t/'
    done

    return 1
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
