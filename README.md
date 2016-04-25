# docker-proxy-acl

A docker unix socket proxy which resticts endpoint usage to allowed sections

## Why?

Exposing `docker.sock` to Docker container comes with security concerns. Depending on what you want
to do from inside the container, the requests can be limited to specific endpoints.

You can enable an endpoint with the `-a` argument. Currently supported endpoints are:

* containers: opens access to `/containers/json` and `/containers/{name}/json`.
* networks: opens access to `/networks` and `/networks/{name}`
* info: opens access to `/info`
* version: opens access to `/version`
* ping: opens access to `/_ping`

To combine arguments, repeat them like this: `./run -a info -a version`.

## Example usage: limiting access from containers

The project [netdata](https://github.com/firehol/netdata) can use the `docker.sock` file to resolve
the container names found in the `cgroups` filesystem, into readable names. Information for this
is only available over the API. Even the `docker` binary uses the Docker API to access this information.

To start a docker-proxy-acl with just the `containers` endpoints:

~~~
./run -a containers
~~~

Using this application, a new socket file is created (`/tmp/docker-proxy-acl/docker.sock`). Specifically
for this example, only the `/containers/json` and `/containers/{name}/json` endpoints are allowed.
This socket file can be passed to the `netdata` container, with an additional option like this:

~~~
-v /tmp/docker-proxy-acl/docker.sock:/var/run/docker.sock
~~~

And now, netdata is free to query `/var/run/docker.sock` from within the container. If netdata is
running on the host, it needs to have access to the same file - but the API stays the same.

## Example usage: exposing limited access over HTTP

Using the same arguments as above, it's possible to provide a limited HTTP endpoint for the API.
To do it, you have to use the separate [docker-proxy](https://github.com/titpetric/docker-proxy) project,
which exposes the docker socket via HTTP. To expose our safe docker socket, use the same `-v` line
from above, to run the new container.

This may be used to provide limited information from Docker hosts to a central monitoring dashboard.

Keep in mind, this might expose some sensitive data:

* The environment passed to docker (may contain passwords, other sensitive data)
* Networking information about containers (ip, gateway, exposed ports)
* Container internals (running commands, process list, source images)

The docker-proxy-acl project doesn't aim to limit the responses in any way. If you're requesting
endponts like `/containters/{name}/json` it will just forward all the response as-is.

> TL;DR - think twice before you're exposing the docker API via HTTP

## TODO

* extend with more ACL rules for other uses/endpoints

