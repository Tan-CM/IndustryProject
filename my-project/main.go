package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	// Note IP is not set here because it should be 0.0.0.0 for container
	// IP with 127.0.0.1 is works for host, but not for container
	log.Fatal(http.ListenAndServe(":8081", nil))

}
