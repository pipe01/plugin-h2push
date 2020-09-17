package plugin_h2push

import (
	"context"
	"fmt"
	"net/http"
)

type Config struct {
	Files []H2PushFile `json:"files,omitempty" toml:"files,omitempty" yaml:"files,omitempty"`
}

type H2PushFile struct {
	URL   string `json:"url" toml:"url" yaml:"url"`
	Match string `json:"match,omitempty" toml:"match,omitempty" yaml:"match,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Files: make([]H2PushFile, 0),
	}
}

type H2Push struct {
	next http.Handler
	cfg  *Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Println("--- creating plugin")
	return &H2Push{next, config}, nil
}

func (h *H2Push) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("--- request on plugin")
	h.next.ServeHTTP(rw, req)
}
