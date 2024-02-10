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

type Story struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Descendants int    `json:"descendants"`
}

const (
	ColorReset   = "\033[0m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorGray    = "\033[37m"
	ColorGreen   = "\033[32m"
	ColorOrange  = "\033[33m"
	Bold         = "\033[1m"
)

func fetchStoriesIDs(storyType string) ([]int, error) {
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

func UserAgentAndStoryTypeHandler(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	class := "Unknown"
	if strings.Contains(ua, "curl") || strings.Contains(ua, "Wget") {
		class = "Curl"
	} else if regexp.MustCompile(`(?i)(firefox|chrome|safari|edge|opera|msie)`).MatchString(ua) {
		class = "Browser"
	}

	storyType := strings.TrimPrefix(r.URL.Path, "/")
	storyIDs, err := fetchStoriesIDs(storyType)
	if err != nil || len(storyIDs) == 0 {
		http.Error(w, "Failed to fetch stories", http.StatusInternalServerError)
		return
	}

	var stories []Story
	for _, id := range storyIDs[:5] {
		story, err := fetchStoryDetails(id)
		if err != nil {
			continue
		}
		stories = append(stories, *story)
	}

	titleCaser := cases.Title(language.English)
	storyTypeTitle := titleCaser.String(storyType)

	var tpl *template.Template
	var tplString string
	if class == "Browser" {
		tplString = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Hacker News {{.StoryType}} Stories</title>
		<style>
			body {
				font-family: 'Arial', sans-serif;
				background-color: #f0f0f0;
				margin: 0;
				padding: 20px;
				color: #333;
			}
			h1 {
				color: #ff6600;
				font-size: 24px;
				margin-bottom: 5px;
			}
			.story-list {
				background-color: #ffffff;
				border: 1px solid #dddddd;
				list-style-type: none;
				padding: 0;
			}
			.story {
				padding: 10px;
				border-bottom: 1px solid #dddddd;
			}
			.story:last-child {
				border-bottom: none;
			}
			.story-title {
				color: #ff6600;
				font-size: 18px;
				margin: 0;
				padding: 0;
			}
			.story-link {
				color: #0066cc;
				text-decoration: none;
			}
			.story-details {
				color: #828282;
				font-size: 14px;
				margin-top: 5px;
			}
			.story-details span {
				margin-right: 10px;
			}
		</style>
	</head>
	<body>
		<h1>Hacker News {{.StoryType}} Stories</h1>
		<ul class="story-list">
			{{range .Stories}}
			<li class="story">
				<p class="story-title"><a class="story-link" href="{{.URL}}">{{.Title}}</a></p>
				<p class="story-details">
					by <span>{{.By}}</span> |
					<span>{{.Descendants}} Comments</span> |
					<span>{{.Score}} Points</span>
				</p>
			</li>
			{{end}}
		</ul>
	</body>
	</html>`
	} else {
		tplString = `Hacker News {{.StoryType}} Stories:
─────────────────────────────────────
{{range .Stories}}
` + Bold + ColorBlue + `{{.Title}}` + ColorReset + `
` + ColorMagenta + `{{.URL}}` + ColorReset + `
` + ColorGray + `by {{.By}}` + ColorReset + ` |  ` + ColorGreen + `{{.Descendants}} Comments` + ColorReset + ` |  ` + ColorOrange + `{{.Score}} Points` + ColorReset + `
─────────────────────────────────────
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

	if class == "Browser" {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Write(buf.Bytes())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/top", UserAgentAndStoryTypeHandler)
	mux.HandleFunc("/new", UserAgentAndStoryTypeHandler)
	mux.HandleFunc("/best", UserAgentAndStoryTypeHandler)

	// Wrap the mux to normalize path
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		mux.ServeHTTP(w, r)
	})

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %s\n", err)
	}
}
