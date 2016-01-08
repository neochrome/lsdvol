[![Build Status](https://travis-ci.org/neochrome/lsdvol.png?branch=master)](https://travis-ci.org/neochrome/lsdvol)

## Usage

### On a Docker host
For a basic listing of volumes in use by the container `my-container`:

```sh
$ lsdvol my-container
```

### From within a running Docker container
The container must have the `/var/run/docker.sock` bound in order for
the program to function properly.

For a basic listing of volumes in use, execute the following from
*within* a Docker container:

```sh
$ lsdvol
```

### Non-standard location of docker.sock
If your `docker.sock` is in a non-standard location, either when bound
in a container or on a Docker host, please specify the location as:

```sh
$ lsdvol --docker-socket /my/other/docker.socket
```

### Other options
For other options, launch with `--help`.
