#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DEFAULT_OUTPUT_PATH="${THIS_DIR}/../output/result.json"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        return 1
    fi

    trap exit SIGINT
    set -e

    cd "${THIS_DIR}"

    # Allow the grader to run locally by changing the output location
    # if not in the docker image.
    local outputPath="${DEFAULT_OUTPUT_PATH}"
    if [[ ! -d $(dirname "${outputPath}") ]] ; then
        outputPath="$(basename "${outputPath}")"
    fi

    # Run grader.
    grade "${outputPath}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to run grader."
        return 3
    fi

    return 0
}

function grade() {
    # Source the student's assignment file.
    source "${THIS_DIR}/assignment.sh"

    echo "Running Bash Grader"

    local score=10
    local message=""

    test_add 1 2 3 "basic" || { score=$((score-2)); message+="Missed test case 'basic'. "; }
    test_add 0 2 2 "one zero" || { score=$((score-2)); message+="Missed test case 'one zero'. "; }
    test_add 0 0 0 "all zero" || { score=$((score-2)); message+="Missed test case 'all zero'. "; }
    test_add -1 2 1 "one negative" || { score=$((score-2)); message+="Missed test case 'one negative'. "; }
    test_add -1 -2 -3 "all negative" || { score=$((score-2)); message+="Missed test case 'all negative'. "; }

    local json_output='{
        "name": "bash",
        "questions": [
            {
                "name": "Task 1: add()",
                "max_points": 10,
                "score": '"$score"',
                "message": "'"$message"'"
            }
        ]
    }'

    echo "$json_output" > "${outputPath}"
}

function test_add() {
    local a=$1
    local b=$2
    local expected=$3
    local feedback=$4

    local result=$(add $a $b)
    [[ $result -eq $expected ]]
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
