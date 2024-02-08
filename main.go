package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func classHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user agent string from the request header
	ua := r.UserAgent()

	// Define the class of the request
	var class string

	// Check if the user agent string contains "curl" or "Wget"
	if strings.Contains(ua, "curl") || strings.Contains(ua, "Wget") {
		// The request is coming from a command-line tool
		class = "Curl"
	} else {
		// Check if the user agent string matches a browser pattern
		browserRegex := regexp.MustCompile(`(?i)(firefox|chrome|safari|edge|opera|msie)`)
		if browserRegex.MatchString(ua) {
			// The request is coming from a browser
			class = "Browser"
		} else {
			// The request is coming from an unknown source
			class = "Unknown"
		}
	}

	// Write the class of the request to the response
	fmt.Fprintln(w, "Your request is coming from a", class)
}

func main() {
	// Create a new ServeMux to handle requests
	mux := http.NewServeMux()

	// Register the classHandler for the /class path
	mux.HandleFunc("/api", classHandler)

	// Start the server on port 8080
	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
