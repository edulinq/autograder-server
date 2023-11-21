#!/bin/bash

# Install a matching version of the Python interface.

readonly TEMP_DIR='/tmp/__autograder__/autograder-py'
readonly REPO_URL='https://github.com/eriq-augustine/autograder-py.git'

function fetch_repo() {
    local branch=$1

    echo "Fetching Python interface repo."

    if [[ -d "${TEMP_DIR}" ]] ; then
        echo "Found existing repo '${TEMP_DIR}', skipping clone."
    else
        mkdir -p "$(dirname "${TEMP_DIR}")"
        git clone "${REPO_URL}" "${TEMP_DIR}"
        if [[ $? -ne 0 ]] ; then
            echo "ERROR: Failed to clone repo."
            return 1
        fi
    fi

    cd "${TEMP_DIR}"

    git checkout "${branch}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to checkout target branch ('${branch}')."
        return 2
    fi

    return 0
}

function install_interface() {
    echo "Installing Python interface."

    cd "${TEMP_DIR}"

    ./install.sh
    return $?
}

function main() {
    if [[ $# -gt 1 ]] ; then
        echo "USAGE: $0 [branch]"
        exit 1
    fi

    local branch=$(git branch --show-current)
    if [[ $# -eq 1 ]] ; then
        branch=$1
    fi

    trap exit SIGINT

    fetch_repo "${branch}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to fetch the Python interface repo."
        return 2
    fi

    install_interface
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to install Python interface."
        return 3
    fi

    echo "Sucessfully installed Python interface."

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
