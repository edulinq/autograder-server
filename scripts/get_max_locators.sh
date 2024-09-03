#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly INTERNAL_DIR="${ROOT_DIR}/internal"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        exit 1
    fi

    trap exit SIGINT

    cd "${ROOT_DIR}"

    local packages_to_check=("api" "procedures" "lms")

    for package in "${packages_to_check[@]}"; do
        for path in "${INTERNAL_DIR}/${package}"/* ; do
            dir=$(basename "${path}")
            parent_dir=$(basename $(dirname "${path}"))

            if [[ ! -d "${path}" ]] ; then
                continue
            fi

            local largestLocator=$(grep -R '("-' "${path}" | sed 's/^.*"\(-[0-9]\{3,4\}\)".*$/\1/' | sort | uniq | tail -n 1)
            # Remove input zero padding (so bash does not think the numebr is octal.
            local cleanLargestLocator=$(echo "${largestLocator}" | sed -E 's/-0+/-/g')
            local nextLocator=$(printf "%04d" "$((cleanLargestLocator - 1))")

            echo -e "Package: $(printf "%-20s" "${parent_dir}/${dir}"),\tMax Locator: $(printf "%-5s" "${largestLocator}"),\tNext Locator: ${nextLocator}"
        done
    done

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
