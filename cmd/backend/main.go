package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 9001, "Backend server port")
	name := flag.String("name", "Backend", "Backend name")

	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s\n", *name)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	address := fmt.Sprintf(":%d", *port)

	log.Printf("%s running on %s", *name, address)

	log.Fatal(http.ListenAndServe(address, nil))
}
