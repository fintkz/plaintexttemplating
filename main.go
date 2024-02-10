package main

import (
	"bytes"
	"fmt"
	"html/template" // Only import this once, used for both HTML and text templates
	"net/http"
	"regexp"
	"strings"
)

// TextTemplateData holds the required data for text template generation
type TextTemplateData struct {
	Ticker    string
	Price     float64
	Changes   map[string]float64
	ChartData []float64
}

// TextTemplate generates the text version of the stock report
func TextTemplate(data TextTemplateData) (string, error) {
	// Define the text template string
	tpl := `Ticker: {{.Ticker}}
Current Price: ${{.Price}}
Changes:
{{- range $period, $change := .Changes}}
  * {{ $period }}: {{ $change | printf "%.2f" }}%
{{- end}}
Chart:
{{- range $price := .ChartData}}
  {{ $price | printf "%-.0f " }}
{{- end}}
`

	// Create a new template
	t := template.Must(template.New("text").Parse(tpl))

	// Execute the template with the data
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	output := buf.String()

	return output, nil
}

// HTMLTemplateData holds the required data for HTML template generation
type HTMLTemplateData struct {
	Ticker    string
	Price     float64
	Changes   map[string]float64
	ChartData []float64
}

// HTMLTemplate generates the HTML version of the stock report
func HTMLTemplate(data HTMLTemplateData) (string, error) {
	// Define the HTML template string
	tpl := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Stock Report</title>
</head>
<body>
<h1>Ticker: {{.Ticker}}</h1>
<p>Current Price: <strong>${{.Price}}</strong></p>
<h2>Changes</h2>
<table>
  <tr>
    <th>Period</th>
    <th>Change</th>
  </tr>
  {{- range $period, $change := .Changes}}
  <tr>
    <td>{{ $period }}</td>
    <td>{{ $change | printf "%.2f" }}%</td>
  </tr>
  {{- end}}
</table>
<h2>Chart</h2>
<pre>
{{- range $price := .ChartData}}
  {{ $price | printf "%-.0f " }}
{{- end}}
</pre>
</body>
</html>`

	// Create a new template
	t := template.Must(template.New("html").Parse(tpl))

	// Execute the template with the data
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	output := buf.String()

	return output, nil
}

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

	// fmt.Fprintln(w, "Your request is coming from a", class)

	var output string
	var err error

	ticker := "NVDA" // Assuming fixed ticker symbol for demo
	price := 290.34  // Assuming fixed price for demo
	changes := map[string]float64{
		"1h":  0.2,
		"24h": 1.5,
		"7d":  3.0,
	}

	switch class {
	case "Curl", "Wget":
		// Generate and respond with a text report
		output, err = TextTemplate(TextTemplateData{
			Ticker:    ticker,
			Price:     price,
			Changes:   changes,
			ChartData: []float64{290.34, 291.00, 289.50}, // Example chart data
		})
		if err != nil {
			http.Error(w, "Failed to generate text report", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, output)

	case "Browser":
		// Generate and respond with an HTML report
		output, err = HTMLTemplate(HTMLTemplateData{
			Ticker:    ticker,
			Price:     price,
			Changes:   changes,
			ChartData: []float64{290.34, 291.00, 289.50}, // Example chart data
		})
		if err != nil {
			http.Error(w, "Failed to generate HTML report", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, output)

	default:
		fmt.Fprintln(w, "Your request is coming from an unknown source")
	}
}

func main() {
	// Set up the HTTP server and route
	http.HandleFunc("/", UserAgentRegexHandler) // Register your handler function

	// Start the HTTP server
	port := ":8080" // Define the port to listen on
	println("Server listening on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err) // Handle any potential errors
	}
}
