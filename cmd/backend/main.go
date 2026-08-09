package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {

	port := flag.String("port", "9001", "Port to run server")
	name := flag.String("name", "Backend A", "Server name")

	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s\n", *name)
	})

	log.Printf("%s running on :%s\n", *name, *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
