# Docker Information

This project uses [Docker](https://www.docker.com/) both
as the mechanism for running isolated graders for each student submission,
and as an optional way to run the server (and other executables) without needing to build/install them.

## Running The Server

The autograder can also be run from a Docker container either
by pulling it from [Docker Hub](https://hub.docker.com/u/edulinq)
or by building the image from source.

### Using the Built Images

We provide images already compiled and hosted on [Github](https://github.com/edulinq/autograder-server/pkgs/container/autograder-server).
These images are built and deployed automatically in this repository's CI.

To run the images, the autograder container must mount two directories:
 - `/var/run/docker.sock` -- The socket that the Docker daemon listens on.
 - `/tmp/autograder-temp/` -- The autograder's temporary directory.

Note that both mounts rely on a POSIX system.

To persist data between container runs,
an additional directory must be mounted: the autograder base/data directory.
When running the autograder directly (not in a Docker container),
this values is set either with the `AUTOGRADER__DIRS__BASE` environmental variable,
or with using the command line option `-c dirs.base='<DATA_DIR>'`.
This directory defaults to the system's [XDG_DATA_HOME](https://specifications.freedesktop.org/basedir-spec/latest/) value.
Below are some defaults based on operating system:
 - Linux -- `~/.local/share`
 - Mac -- `~/Library/Application Support`

### Running with the Docker CLI

To run the container run:
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    edulinq/autograder-server <command>
```

Where `<command>` can be any command form the `cmd` folder.
For example, you can run `version` using:
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    edulinq/autograder-server version
```

If you want to run the server, it could be useful to add the -p flag to the command, as shown below:
```
-p <host port>:<container port>
```

For example, you may use the following command (which uses the autograder's default port of 8080):
```
docker run -it --rm \
    -p 8080:8080 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    edulinq/autograder-server server
```

To ensure your changes persist between container runs,
make sure that you mount your data directory.
For example if you are on Linux,
you could use the following command:
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    -v $(realpath ~/.local/share):/data \
    edulinq/autograder-server version
```

Note the use of `realpath`, because `docker run` requires absolute paths.

### Running with the Python Script

We provide a convenience script that runs the container using the same Docker CLI as above,
but has the default values builtin.
This script lives at [docker/run-docker-server.py](../docker/run-docker-server.py),
and will use all the defaults mentioned above automatically.

For example, to run the above `version` command, you can use:
```
./docker/run-docker-server.py version
```

Use `--help` to see all the functionality of the script:
```
./docker/run-docker-server.py --help
```

### Building Images from Source

To build the images, you can use the following commands from the repository's root directory:
```
docker build -f docker/Dockerfile -t my-autograder-server .
```

Note the `my-` prefix that was added to the image tags to indicate that you built them.
