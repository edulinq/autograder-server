#!/bin/bash

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

    echo "API Locators:"

    for path in "${API_DIR}"/* ; do
        dir=$(basename "${path}")

        if [[ ! -d "${path}" ]] ; then
            continue
        fi

        local largestLocator=$(grep -R '("-' "${path}" | sed 's/^.*"\(-[0-9]\{3\}\)".*$/\1/' | sort | uniq | tail -n 1)
        # Remove input zero padding (so bash does not think the numebr is octal.
        local cleanLargestLocator=$(echo "${largestLocator}" | sed -E 's/-0+/-/g')
        local nextLocator=$(printf "%04d" "$((cleanLargestLocator - 1))")

        echo -e "Package: $(printf "%-12s" "${dir}"),\tMax Locator: $(printf "%-4s" "${largestLocator}"),\tNext Locator: ${nextLocator}"
    done

    echo ""
    echo "Procedure Locators:"

    for path in "${PROCEDURES_DIR}"/* ; do
        dir=$(basename "${path}")

        if [[ ! -d "${path}" ]] ; then
            continue
        fi

        local largestLocator=$(grep -R '("-' "${path}" | sed 's/^.*"\(-[0-9]\{4\}\)".*$/\1/' | sort | uniq | tail -n 1)
        # Remove input zero padding (so bash does not think the numebr is octal.
        local cleanLargestLocator=$(echo "${largestLocator}" | sed -E 's/-0+/-/g')
        local nextLocator=$(printf "%04d" "$((cleanLargestLocator - 1))")

        echo -e "Package: $(printf "%-12s" "${dir}"),\tMax Locator: $(printf "%-4s" "${largestLocator}"),\tNext Locator: ${nextLocator}"
    done

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
