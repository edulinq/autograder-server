#!/bin/bash 

readonly DEFAULT_PORT=8080
readonly TEMP_DIR="/tmp/autograder-temp"
readonly IMAGENAME="autograder"

# TODO(Batu): Figure out port.
# TODO(Batu): Figure out /tmp/autograder-temp.

function main() {
    if [[ $# -eq 0 ]] ; then
        echo "Usage: $0 <autograder command>"
        exit 1
    fi

    set -e
    trap exit SIGINT

    mkdir -p "${TEMP_DIR}"

    docker run -it --rm -p "${DEFAULT_PORT}:${DEFAULT_PORT}" -v /var/run/docker.sock:/var/run/docker.sock -v "${TEMP_DIR}:${TEMP_DIR}" "${IMAGENAME}" $@
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"