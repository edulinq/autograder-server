#!/bin/bash

sudo setcap CAP_NET_BIND_SERVICE=+eip /scratch/eriq/code/autograder/autograder/bin/server
