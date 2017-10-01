# Small POC

This project aims to show a minimal working private registry in a multi-node docker swarm.


## Setup

The proejct is uses [`vndr`](https://github.com/LK4D4/vndr), run

```sh
$ go get github.com/LK4D4/vndr
```
to intall vndr.  The project is run via the following commands:

```sh
vndr
./swarm init
eval $(docker-machine env manager1)
./swarm build
eval $(docker-machine env worker1)
./swarm build
eval $(docker-machine env manager1)
docker pull registry:2
./swarm deploy
```

Assuming that you said yes during the init, visit `http://builderpoc:9090/` and
see it should say

```
Hello, this is the function build poc
```

To test the "build" and push to a private registry, visit `http://builderpoc:9090/build`

To stop everything use

```sh
./swarm stop
```

To remove everything

```sh
./swarm teardown
```

## Current state
The project currently fails to push images to the private repository because the `server` container DNS resolution does not  resolve the name of the `privateregistry` service.  You can see the dns failure as follows

```sh
eval $(docker-machine env worker1)
docker exec $(docker ps --filter "name=builderpoc_server" -q) dig privateregistry
```

You can validate that the registry is running by using

```sh
curl -k https://builderpoc:5001/v2/
```

When the build is working, we expect a non-empty and non-error response from

```sh
curl -k https://builderpoc:5001/v2/_catalog
curl -k https://builderpoc:5001/v2/privatefunc/wordcount/manifests/latest
```