package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"

	"github.com/gorilla/mux"
)

var env Env

func main() {
	env = Env{}
	flag.StringVar(&env.Host, "host", env.Host, "Set the host")
	flag.StringVar(&env.Port, "port", env.Port, "Set the port")
	flag.StringVar(&env.Registry, "registry", env.Registry, "Set the host name of the private registry used to store docker images for the functions.")

	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", hello)
	router.HandleFunc("/build", buildFunction)

	address := fmt.Sprintf("%s:%s", env.Host, env.Port)
	log.Printf("Starting server at %s\n", address)
	http.ListenAndServe(address, router)
}

type Env struct {
	Host     string
	Port     string
	Registry string
	Gateway  string
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Println("Request to hello")
	fmt.Fprintf(w, "Hello, this is the function build poc")
}

func buildFunction(w http.ResponseWriter, r *http.Request) {
	log.Println("Request to buildFunction")
	// pull and then push the functions/wordcount image to the private repo
	source := "functions/wordcount"
	target := fmt.Sprintf("%s/privatefunc/wordcount", env.Registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dockerclient, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	// empty auth
	auth := createRegistryAuth("", "", env.Registry)
	pullOpts := types.ImagePullOptions{
		RegistryAuth: auth,
	}
	log.Println("Starting Image pull")
	response, err := dockerclient.ImagePull(ctx, source, pullOpts)
	if err != nil {
		fmt.Fprintf(w, "Image pull failed with:\n\t%s", err)
		log.Println(err)
		return
	}
	defer response.Close()

	if err := jsonmessage.DisplayJSONMessagesStream(response, os.Stdout, os.Stdout.Fd(), true, nil); err != nil {
		fmt.Fprintf(w, "Image pull failed with:\n\t%s", err)
		log.Println(err)
		return
	}

	log.Printf("Tagging image as %s\n", target)
	err = dockerclient.ImageTag(ctx, source, target)
	if err != nil {
		fmt.Fprint(w, "Image tag failed")
		log.Println(err)
		return
	}

	log.Println("Starting Image push")
	pushOpts := types.ImagePushOptions{
		RegistryAuth: auth, // RegistryAuth is the base64 encoded credentials for the registry
	}
	response, err = dockerclient.ImagePush(ctx, target, pushOpts)
	if err != nil {
		fmt.Fprintf(w, "Push to private repo failed %s", err)
		log.Println(err)
		return
	}
	defer response.Close()

	if err := jsonmessage.DisplayJSONMessagesStream(response, os.Stdout, os.Stdout.Fd(), true, nil); err != nil {
		fmt.Fprintf(w, "Image push failed due to:\n%s", err)
		log.Println(err)
		return
	}
	fmt.Fprintf(w, "Push to private repo complete")
}

func createRegistryAuth(email string, pass string, server string) string {
	authJSON := fmt.Sprintf(`
	{
	  "username": "%s",
	  "password": "%s",
	  "email": "%s",
	  "serveraddress": "%s"
	}`, email, pass, email, server)

	return base64.StdEncoding.EncodeToString([]byte(authJSON))
}
