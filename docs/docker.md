# Docker Information

This project uses [Docker](https://www.docker.com/) both
as the mechanism for running isolated graders for each student submission,
and as an optional way to run the server (and other executables) without needing to build/install them.

## Running The Server

The autograder can also be run from a Docker container either
by pulling it from [Docker Hub](https://hub.docker.com/u/edulinq)
or by building the image from source.

### Using the Built Images

We provide two images already compiled and hosted on [Docker Hub](https://hub.docker.com/u/edulinq):
 - `edulinq/autograder-server-prebuilt` --
   An image that contains all the autograder's functionality and executables that have already been built.
 - `edulinq/autograder-server-slim` --
   All the same functionality of the `prebuilt` version, but with a substantially smaller size.
   To allow for a smaller size, the target executable will be rebuilt from source when the container is run.

These images are built and deployed automatically in this repository's CI.
The latest version will always reflect the most recent passing version of the `main` branch of this repo.
For any example code we will be using the prebuilt image, but either image will work.

To run the images, the autograder container requires two mounts:
 - `/var/run/docker.sock` -- The socket that the Docker daemon listens on.
 - `/tmp/autograder-temp/` -- The autograder's temporary directory.

Note that both mounts rely on a POSIX system.

### Running with the Docker CLI

To run the container run:
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    edulinq/autograder-server-prebuilt <command>
```

Where `<command>` can be any command form the `cmd` folder.
For example, you can run `version` using:
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp/autograder-temp/:/tmp/autograder-temp \
    edulinq/autograder-server-prebuilt version
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
    edulinq/autograder-server-prebuilt server
```

### Running with the Python Script

We provide a convenience script that run the container using the same Docker CLI as above,
but has the default values builtin.
This script lives at [docker/run-docker-server.py](docker/run-docker-server.py),
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
# Prebuilt Image
docker build -f docker/prebuilt/Dockerfile -t my-autograder-server-prebuilt .

# Slim Image
docker build -f docker/prebuilt/Dockerfile -t my-autograder-server-prebuilt .
```

Note the `my-` prefix that was added to the image tags to indicate that you built them.
