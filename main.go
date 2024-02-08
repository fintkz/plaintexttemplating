package main

import (
	"fmt"
	"log"
	"net/http"
)

func useragentHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user agent string from the request header
	ua := r.Header.Get("User-Agent")

	// Write the user agent string to the response
	fmt.Fprintln(w, "Your user agent is:", ua)
}

func main() {
	// Create a new ServeMux to handle requests
	mux := http.NewServeMux()

	// Register the useragentHandler for the /useragent path
	mux.HandleFunc("/useragent", useragentHandler)

	// Start the server on port 8080
	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
