# Hacker News Story Fetcher

This Go project fetches the top 5 stories from Hacker News based on the specified story type (top, new, or best) and displays them in a user-friendly format. The project provides a simple web server that detects the user agent and renders the stories accordingly.

## Features

- Fetches the top 5 stories from Hacker News based on the specified story type
- Detects the user agent (browser or command-line tool) and renders the stories in a suitable format
- Displays the story title, URL, author, number of comments, and points

## Installation

1. Clone the repository:

  git clone https://github.com/fintkz/plaintexttemplating.git

2. Navigate to the project directory:

  cd plaintexttemplating

3. Install the dependencies:

  go mod download

## Usage

### For Linux and macOS

1. Build the project:

  go build

2. Run the server:

  ./plaintexttemplating

3. Open a web browser or use a command-line tool like `curl` to access the stories:

  - For top stories: `http://localhost:8080/top`
  - For new stories: `http://localhost:8080/new`
  - For best stories: `http://localhost:8080/best`

### For Windows

1. Build the project:

  go build

2. Run the server:

  plaintexttemplating.exe

3. Open a web browser or use a command-line tool like `curl` to access the stories:

  - For top stories: `http://localhost:8080/top`
  - For new stories: `http://localhost:8080/new`
  - For best stories: `http://localhost:8080/best`

## How it Works

1. The project starts a web server and listens on port 8080.
2. When a request is made to `/top`, `/new`, or `/best`, the corresponding handler function is invoked.
3. The handler function fetches the story IDs for the specified story type from the Hacker News API.
4. It then fetches the details of the top 5 stories using the story IDs.
5. Based on the user agent, the handler function renders the stories in a suitable format:
  - For browsers, it generates an HTML page with the stories and their details.
  - For command-line tools, it generates a plain text output with the stories and their details.
6. The rendered content is sent back to the client as the response.

## TODOs

- Add error handling and logging for better debugging and monitoring
- Implement pagination to allow fetching more than 5 stories
- Add support for filtering stories based on additional criteria (e.g., by author, score, etc.)
- Improve the HTML template for better responsiveness and mobile compatibility
- Add unit tests to ensure the correctness of the functionality
- Consider adding a caching mechanism to reduce the number of API requests

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).