package main

import (
	"log"
	"net/http"
	"os"

	api "s2-leads-api/api"
)

func main() {
	http.HandleFunc("/", api.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
