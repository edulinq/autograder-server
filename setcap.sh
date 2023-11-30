#!/bin/bash

sudo setcap CAP_NET_BIND_SERVICE=+eip $(realpath bin/server)
