package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// fetchStoriesIDs fetches story IDs from Hacker News based on the story type.
func fetchStoriesIDs(storyType string) ([]int, error) {
	// Correctly construct the URL for each story type
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/%sstories.json?print=pretty", storyType)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

// UserAgentAndStoryTypeHandler determines the client's user agent and serves the appropriate story type.
func UserAgentAndStoryTypeHandler(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	class := "Unknown"
	if strings.Contains(ua, "curl") || strings.Contains(ua, "Wget") {
		class = "Curl"
	} else if regexp.MustCompile(`(?i)(firefox|chrome|safari|edge|opera|msie)`).MatchString(ua) {
		class = "Browser"
	}

	// Determine story type from URL path
	storyType := strings.TrimPrefix(r.URL.Path, "/")
	storyType = strings.TrimSuffix(storyType, "/") // Ensure no trailing slash
	if storyType != "top" && storyType != "new" && storyType != "best" {
		http.Error(w, "Invalid story type", http.StatusBadRequest)
		return
	}

	// Fetch story IDs based on the type
	storyIDs, err := fetchStoriesIDs(storyType)
	if err != nil || len(storyIDs) == 0 {
		http.Error(w, "Failed to fetch stories", http.StatusInternalServerError)
		return
	}

	// Fetch details for the top 10 stories of the specified type
	var stories []Story
	for _, id := range storyIDs[:5] { // Limiting to top 10 stories for brevity
		story, err := fetchStoryDetails(id)
		if err != nil {
			continue // Skip stories that fail to fetch
		}
		stories = append(stories, *story)
	}

	titleCaser := cases.Title(language.English)
	storyTypeTitle := titleCaser.String(storyType) // Use this for your template data

	// Prepare the template based on the class
	var tpl *template.Template
	var tplString string
	if class == "Browser" {
		tplString = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Hacker News {{.StoryType}} Stories</title>
</head>
<body>
<h1>Hacker News {{.StoryType}} Stories</h1>
<ul>
{{range .Stories}}
	<li><a href="{{.URL}}">{{.Title}}</a> by {{.By}}</li>
{{end}}
</ul>
</body>
</html>
`
	} else {
		tplString = `Hacker News {{.StoryType}} Stories:
{{range .Stories}}
- Title: {{.Title}}
  URL: {{.URL}}
  By: {{.By}}
{{end}}
`
	}

	tpl = template.Must(template.New("webpage").Parse(tplString))
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]interface{}{
		"StoryType": storyTypeTitle,
		"Stories":   stories,
	})
	if err != nil {
		http.Error(w, "Failed to generate content", http.StatusInternalServerError)
		return
	}

	// Set the appropriate content type
	if class == "Browser" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	// Write the output
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/top", UserAgentAndStoryTypeHandler)
	http.HandleFunc("/new", UserAgentAndStoryTypeHandler)
	http.HandleFunc("/best", UserAgentAndStoryTypeHandler)

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %s\n", err)
	}
}
