#!/bin/bash

# Check for duplicate API and procedure locators.
# This does not work for locators that are on a different line from the locator creation function call.
# Specifically, we will look for 'Error("-' to indicate where API locators are defined.
# Specifically, we will look for 'Error("-' to indicate where procedure locators are defined.

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly API_DIR="${ROOT_DIR}/internal/api"
readonly PROCEDURES_DIR="${ROOT_DIR}/internal/procedures"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    local duplicateAPILocators=$(grep -R 'Error("-' "${API_DIR}" | sed 's/^.*"\(-[0-9]\{3,\}\)".*$/\1/' | sort | uniq -c | grep -ZvE '^\s+1\s+' | sed 's/^\s*\([0-9]\+\)\s\+\(-[0-9]\{3,\}\)$/\2/')

    local duplicateProceduresLocators=$(grep -R 'Error("-' "${PROCEDURES_DIR}" | sed 's/^.*"\(-[0-9]\{3,\}\)".*$/\1/' | sort | uniq -c | grep -ZvE '^\s+1\s+' | sed 's/^\s*\([0-9]\+\)\s\+\(-[0-9]\{3,\}\)$/\2/')

    if [[ -z ${duplicateAPILocators} ]] && [[ -z ${duplicateProceduresLocators} ]]; then
        echo "No duplicate locators found."
        return 0
    fi

    echo "Found duplicate locators."

    for duplicateAPILocator in ${duplicateAPILocators} ; do
        echo "${duplicateAPILocator}:"
        grep -R --line-number "Error(\"${duplicateAPILocator}\"" "${API_DIR}" | sed 's/^/\t/'
    done

    for duplicateProceduresLocator in ${duplicateProceduresLocators} ; do
        echo "${duplicateProceduresLocator}:"
        grep -R --line-number "Error(\"${duplicateProceduresLocator}\"" "${PROCEDURES_DIR}" | sed 's/^/\t/'
    done

    return 1
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
