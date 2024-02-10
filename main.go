package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// Story represents a Hacker News story.
type Story struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Descendants int    `json:"descendants"` // Number of comments
}

// fetchTopStoriesIDs fetches the top stories IDs from Hacker News.
func fetchTopStoriesIDs() ([]int, error) {
	resp, err := http.Get("https://hacker-news.firebaseio.com/v0/topstories.json?print=pretty")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// fetchStoryDetails fetches the details of a story by its ID.
func fetchStoryDetails(id int) (*Story, error) {
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var story Story
	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return nil, err
	}

	return &story, nil
}

// UserAgentRegexHandler determines the client's user agent and serves the content accordingly.
func UserAgentRegexHandler(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()

	var class string
	if strings.Contains(ua, "curl") || strings.Contains(ua, "Wget") {
		class = "Curl"
	} else if regexp.MustCompile(`(?i)(firefox|chrome|safari|edge|opera|msie)`).MatchString(ua) {
		class = "Browser"
	} else {
		class = "Unknown"
	}

	// Fetch top story IDs
	topStoryIDs, err := fetchTopStoriesIDs()
	if err != nil || len(topStoryIDs) == 0 {
		http.Error(w, "Failed to fetch top stories", http.StatusInternalServerError)
		return
	}

	// Fetch details for the top 10 stories
	var stories []Story
	for _, id := range topStoryIDs[:10] { // Limiting to top 10 stories for brevity
		story, err := fetchStoryDetails(id)
		if err != nil {
			continue // Skip stories that fail to fetch
		}
		stories = append(stories, *story)
	}

	// Choose the template based on the user agent
	if class == "Browser" {
		htmlTpl := template.Must(template.New("html").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Hacker News Top Stories</title>
</head>
<body>
<h1>Hacker News Top Stories</h1>
<ul>
{{range .}}
    <li><a href="{{.URL}}">{{.Title}}</a> by {{.By}}</li>
{{end}}
</ul>
</body>
</html>
`))
		var buf bytes.Buffer
		if err := htmlTpl.Execute(&buf, stories); err != nil {
			http.Error(w, "Failed to generate HTML content", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, buf.String())
	} else if class == "Curl" {
		textTpl := template.Must(template.New("text").Parse(`
Hacker News Top Stories:
{{range .}}
- Title: {{.Title}}
  URL: {{.URL}}
  By: {{.By}}
{{end}}
`))
		var buf bytes.Buffer
		if err := textTpl.Execute(&buf, stories); err != nil {
			http.Error(w, "Failed to generate text content", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, buf.String())
	} else {
		http.Error(w, "Unsupported client", http.StatusBadRequest)
	}
}

func main() {
	http.HandleFunc("/", UserAgentRegexHandler) // Register the handler function

	// Start the HTTP server on port 8080
	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
