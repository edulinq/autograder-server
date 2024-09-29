#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly INTERNAL_DIR="${ROOT_DIR}/internal"

function process_directory() {
    local path="$1"

    local dirname=$(basename "${path}")
    local parent_dirname=$(basename "$(dirname "${path}")")

    local largestLocator=$(grep -RoEhI '("\-[0-9]{3,4}")' "${path}" 2>/dev/null | sed 's/^.*"\(-[0-9]\{3,4\}\)".*$/\1/' | sort | uniq | tail -n 1)

    if [[ -z "${largestLocator}" ]] ; then
        return 0
    fi

    # Remove input zero padding (so bash does not think the number is octal).
    local cleanLargestLocator=$(echo "${largestLocator}" | sed -E 's/-0+/-/g')
    local nextLocator=$(printf "%04d" "$((cleanLargestLocator - 1))")

    echo -e "Package: $(printf "%-20s" "${parent_dirname}/${dirname}"),\tMax Locator: $(printf "%-5s" "${largestLocator}"),\tNext Locator: ${nextLocator}"

    return 1
}

function main() {
    if [[ $# -ne 0 ]] ; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    for dir in "${INTERNAL_DIR}"/* ; do
        local found=0
        for path in "${dir}"/* ; do
            if [[ -d "${path}" ]] ; then
                process_directory "${path}"
                (( found += $? ))
            fi
        done

        if [[ ${found} -eq 0 ]] ; then
            process_directory "${dir}"
        fi
    done

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
