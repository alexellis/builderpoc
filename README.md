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

The project current works as needed, although not as expected. Immediately after
initializing and deploying the project you can validate that the registry is running by using

```sh
$ curl -k https://builderpoc:5001/v2/_catalog
{"repositories":[]}
```
This shows that the registry is currently empty.

You can push the test image into the registry using the build endpoint

```sh
$ curl http://builderpoc:9090/build
Push to private repo complete
```
This may take a moment becaue it needs to pull and then push an image.

Finally, you can valdiate that the new image is in the repo using

```sh
curl -k https://builderpoc:5001/v2/_catalog
{"repositories":["privatefunc/wordcount"]}
```

and `curl -k https://builderpoc:5001/v2/privatefunc/wordcount/manifests/latest`
to see the actual content of the latest tag.


All of the above is expected.  What is/was not expected is the configuration required for this to work. In Docker Swarm it appears that the services are locally bound to each other.  Specifically, while we defined the registry as the `privateregistry` in the docker compose file, it is actually exposed on localhost.  Thus, the build server command is

```yaml
command: [
    "builder",
    "-port=9090",
    "-host=0.0.0.0",
    "-registry=localhost:5001"
]
```

not

```yaml
command: [
    "builder",
    "-port=9090",
    "-host=0.0.0.0",
    "-registry=privateregistry:5001"
]
```

which I was expecting.


With this in place, you can verify that the OpenFaaS gateway can access and deploy this image using

```sh
curl -H "Content-Type: application/json" -X POST -d '{"service":"wc","network":"builderpoc_functions", "image": "localhost:5001/privatefunc/wordcount"}' http://builderpoc:8080/system/functions
```


and then running the function with

```sh
$ curl 'http://builderpoc:8080/function/wc' -H 'Content-Type: text/plain' -H 'Accept: application/json, text/plain, */*' --data-binary 'hi, this is just a test'
        0         6        23
```