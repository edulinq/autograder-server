#!/usr/bin/env python3

"""
Run the autograder server (or other commands) using an existing Docker image.

Possible commands are located in the `cmd` directory of this repository.

To have results of this script persist outside of the container,
you will have to ensure that the data directory (set via --data-dir) is set correctly.
If you do not set this value,
then it will be computed by first checking the environmental variable 'AUTOGRADER__DIRS__BASE'
and then by guessing based on the operating system.

Note that this script is only meant to be used by POSIX systems.
"""

import argparse
import os
import shlex
import signal
import subprocess
import sys
import time

THIS_DIR = os.path.abspath(os.path.dirname(os.path.realpath(__file__)))
CMD_DIR = os.path.join(THIS_DIR, '..', 'cmd')

DEFAULT_DOCKER_IMAGE = 'edulinq/autograder-server-prebuilt:latest'

DOCKER_CONTAINER_DOCKER_SOCKET = '/var/run/docker.sock'
DEFAULT_HOST_DOCKER_SOCKET = '/var/run/docker.sock'

DOCKER_CONTAINER_TEMP_DIR = '/tmp/autograder-temp'
DEFAULT_HOST_TEMP_DIR = '/tmp/autograder-temp'

DOCKER_CONTAINER_DATA_DIR = '/data'
# Map of sys.platform to default data dir.
DEFAULT_HOST_DATA_DIRS = {
    'darwin': '~/Library/Application Support',
    'linux': '~/.local/share',
}
DATA_DIR_ENV_VARIABLE = 'AUTOGRADER__DIRS__BASE'

DOCKER_CONTAINER_PORT = 8080
DEFAULT_HOST_PORT = 8080

KILL_SLEEP_TIME_SECS = 1.0

def main(args):
    command = [
        'docker', 'run',
        '--rm',
        '--interactive', '--tty',
        '--volume', '%s:%s' % (_normlaize_path(args.socket), DOCKER_CONTAINER_DOCKER_SOCKET),
        '--volume', '%s:%s' % (_normlaize_path(args.temp_dir), DOCKER_CONTAINER_TEMP_DIR),
    ]

    for raw_mount in args.mount:
        host_path, container_path = _parse_mount(raw_mount)
        command += [
            '--volume', '%s:%s' % (host_path, container_path),
        ]

    if (not args.no_data):
        if (args.data_dir is None):
            print("Could not find data directory for this system, please specify with --data-dir.", file = sys.stderr)
            return 1

        command += [
            '--volume', '%s:%s' % (_normlaize_path(args.data_dir), DOCKER_CONTAINER_DATA_DIR),
        ]

    if (not args.no_port):
        command += [
            '--publish', '%d:%d' % (args.port, DOCKER_CONTAINER_PORT),
        ]

    command += args.docker_args
    command += [
        args.image,
        args.command,
    ]
    command += args.command_args

    if (args.echo):
        print(shlex.join(command))

    result = _run(command)
    return result

def _parse_mount(raw_mount):
    parts = raw_mount.split(':')
    if (len(parts) != 2):
        raise ValueError("Could not parse mount paths (exactly one colon not found): '%s'." % (raw_mount))

    host_path = _normlaize_path(parts[0])
    if (not os.path.isdir(host_path)):
        raise ValueError("Host mount path does not exist or is not a directory: '%s'." % (host_path))

    container_path = parts[1]
    if (not os.path.isabs(container_path)):
        raise ValueError("Container mount path is not absolute: '%s'." % (container_path))

    return host_path, container_path

def _normlaize_path(path):
    return os.path.abspath(os.path.expanduser(os.path.expandvars(path)))

def _run(args):
    try:
        process = subprocess.Popen(args)
        return process.wait()
    except KeyboardInterrupt:
        # Try to end the process gracefully.
        process.send_signal(signal.SIGINT)

        try:
            process.wait(KILL_SLEEP_TIME_SECS)
        except subprocess.TimeoutExpired:
            # End the process hard.
            process.send_signal(signal.SIGKILL)
            process.terminate()
            process.kill()
            time.sleep(KILL_SLEEP_TIME_SECS)

    if (process.returncode is not None):
        return process.returncode

    return 99

def _discover_cmds():
    if (not os.path.isdir(CMD_DIR)):
        raise ValueError("Command directory ('%s') either does not exist or is not a directory." % (CMD_DIR))

    cmds = []
    for dirent in os.listdir(CMD_DIR):
        cmds.append(dirent)

    return list(sorted(cmds))

def _compute_default_data_dir():
    # Check for an environmental variable.
    data_dir = os.environ.get(DATA_DIR_ENV_VARIABLE, None)
    if (data_dir is not None):
        return data_dir

    # To avoid additional dependencies for this script,
    # we will guess differnet directories instead of using a more robust library.
    return DEFAULT_HOST_DATA_DIRS.get(sys.platform, None)

def _load_args():
    parser = argparse.ArgumentParser(
            description = __doc__.strip(),
            formatter_class = argparse.RawTextHelpFormatter)

    # Find all the possible CMDs.
    cmds = _discover_cmds()

    # Compute the default data directory.
    default_host_data_dir = _compute_default_data_dir()

    parser.add_argument('command', metavar = 'COMMAND',
        action = 'store', type = str,
        choices = cmds,
        help = 'The command to execute. Known commands: %s.' % (cmds))

    parser.add_argument('command_args', metavar = 'COMMAND_ARGUMENTS',
        action = 'store', type = str, nargs = '*', default = [],
        help = 'Arguments passed to the command you are running.')

    parser.add_argument('--image', dest = 'image',
        action = 'store', type = str, default = DEFAULT_DOCKER_IMAGE,
        help = 'The docker image to use (default: "%(default)s").')

    parser.add_argument('--data-dir', dest = 'data_dir',
        action = 'store', type = str, default = default_host_data_dir,
        help = 'The location of the autograder\'s base/data dir (default: "%(default)s").')

    parser.add_argument('--socket', dest = 'socket',
        action = 'store', type = str, default = DEFAULT_HOST_DOCKER_SOCKET,
        help = 'The location of the docker daemon\'s socket (default: "%(default)s").')

    parser.add_argument('--temp-dir', dest = 'temp_dir',
        action = 'store', type = str, default = DEFAULT_HOST_TEMP_DIR,
        help = 'The location of the autograder\'s temp dir (default: "%(default)s").')

    parser.add_argument('--port', dest = 'port',
        action = 'store', type = int, default = DEFAULT_HOST_PORT,
        help = 'The port to use (default: "%(default)s").')

    parser.add_argument('--mount', dest = 'mount',
        action = 'append', type = str, default = [],
        help = ('Specify a directory to mount in the Docker container.'
            + ' Argument must be formatted as "<host dir>:<container dir>",'
            + ' e.g., "my-host-dir:/tmp/my-container-dir".'
            + ' Paths cannot have a colon (":") in them.'
            + ' The host directory can be relative, and the container dir must be absolute.'
            + ' Can be specified multiple times.'))

    parser.add_argument('-d', '--docker-arg', dest = 'docker_args',
        action = 'append', type = str, default = [],
        help = 'Additional arguments to pass to docker.')

    parser.add_argument('--no-data', dest = 'no_data',
        action = 'store_true', default = False,
        help = "Do not mount any data directory (default: %(default)s).")

    parser.add_argument('--no-port', dest = 'no_port',
        action = 'store_true', default = False,
        help = "Do not open a port for this container (default: %(default)s).")

    parser.add_argument('--echo', dest = 'echo',
        action = 'store_true', default = False,
        help = "Echo the docker command before running (default: %(default)s).")

    return parser.parse_args()

if (__name__ == '__main__'):
    sys.exit(main(_load_args()))
