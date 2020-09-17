package plugin_h2push

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	linkRegex        = regexp.MustCompile(`(?m)<([^>]+)>;\s+rel="?(\w+)"?;\s+as="?(\w+)"?`)
	absoluteURLRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z\d+\-.]*:`)
)

type Config struct {
	Files []H2PushFile `json:"files,omitempty" toml:"files,omitempty" yaml:"files,omitempty"`
}

type H2PushFile struct {
	URL   string `json:"url" toml:"url" yaml:"url"`
	Match string `json:"match,omitempty" toml:"match,omitempty" yaml:"match,omitempty"`
}

type pushFile struct {
	H2PushFile
	MatchRegexp *regexp.Regexp
}

type linkHeader struct {
	FileName, Rel, Kind string
}

func CreateConfig() *Config {
	return &Config{
		Files: make([]H2PushFile, 0),
	}
}

type H2Push struct {
	next  http.Handler
	files []pushFile
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	files := make([]pushFile, len(config.Files))

	fmt.Println("hello")

	var err error
	for i, f := range config.Files {
		compiled := pushFile{f, nil}

		if f.Match != "" {
			compiled.MatchRegexp, err = regexp.Compile(f.Match)

			if err != nil {
				return nil, fmt.Errorf("failed to compile regexp %q: %w", f.Match, err)
			}
		}

		files[i] = compiled
	}

	return &H2Push{next, files}, nil
}

func (h *H2Push) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("--- request on plugin")

	pusher, isPushable := rw.(http.Pusher)

	if isPushable && h.files != nil && len(h.files) > 0 {
		pushFiles(pusher, h.files, req)
	}

	h.next.ServeHTTP(rw, req)

	if isPushable {
		pushLinks(pusher, rw.Header()["Link"])
	}
}

func pushFiles(p http.Pusher, files []pushFile, req *http.Request) {
	for _, file := range files {
		if file.MatchRegexp != nil && !file.MatchRegexp.MatchString(req.URL.Path) {
			continue
		}

		p.Push(file.URL, nil)
	}
}

func pushLinks(p http.Pusher, linkHeaders []string) {
	for _, h := range linkHeaders {
		link, err := parseLink(h)
		if err != nil {
			log.Printf("failed to parse Link header: %w", err)
			return
		}

		fname := normalizePath(link.FileName)

		p.Push(fname, nil)
	}
}

func parseLink(l string) (*linkHeader, error) {
	groups := linkRegex.FindStringSubmatch(l)

	if len(groups) != 4 {
		return nil, fmt.Errorf("invalid link header %q", l)
	}

	return &linkHeader{groups[1], groups[2], groups[3]}, nil
}

// prepends "/" to the path if it doesn't already and it isn't an absolute URL (i.e. "http://google.com/file.txt")
func normalizePath(path string) (absolutePath string) {
	if !strings.HasPrefix(path, "/") && !absoluteURLRegex.MatchString(path) {
		return "/" + path
	}

	return path
}
