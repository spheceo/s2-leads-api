package main

import (
	"log"
	"net/http"

	api "s2-leads-api/api"
)

func main() {
	http.HandleFunc("/", api.Handler)

	log.Println("🎉 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}