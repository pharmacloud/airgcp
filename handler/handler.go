package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.Printf("GOOGLE_CLOUD_PROJECT: %s", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	log.Printf("DATABASE: %s", os.Getenv("DATABASE"))
	log.Printf("PARAM_AUTH: %s", os.Getenv("PARAM_AUTH"))
	log.Printf("MAPKEY: %s", os.Getenv("MAPKEY"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(req http.ResponseWriter, res *http.Request) {
		log.Printf("!!!!3")
	})
	http.ListenAndServe(":18080", mux)
}
