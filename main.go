package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	httpServe := http.Server{Handler: serveMux, Addr: ":8080"}
	if err := httpServe.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
