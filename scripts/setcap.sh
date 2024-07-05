#!/bin/bash

readonly THIS_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd | xargs realpath)"
readonly ROOT_DIR="${THIS_DIR}/.."
readonly SERVER_BIN_PATH="${ROOT_DIR}/bin/server"

sudo setcap CAP_NET_BIND_SERVICE=+eip $(realpath "${SERVER_BIN_PATH}")
