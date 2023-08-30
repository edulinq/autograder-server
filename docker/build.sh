#!/bin/bash

# Build all the docker images in this directory.
# The images are build in lexicographic order.

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly BASE_IMAGE_NAME='autograder'

function main() {
    trap exit SIGINT
    set -e

    for dockerfile in $(ls "${THIS_DIR}"/*/Dockerfile | sort) ; do
        local buildDir=$(dirname "${dockerfile}")
        local subImageName=$(basename "${buildDir}" | sed 's/^[0-9]\+-//')

        echo "Building '${BASE_IMAGE_NAME}.${subImageName}' ..."
        docker build --no-cache --tag "${BASE_IMAGE_NAME}.${subImageName}" --file "${dockerfile}" "${buildDir}" $@
    done

    exit 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
