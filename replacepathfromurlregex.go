// Package replacepathfromurlregex a treafik plugin.
package replacepathfromurlregex

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	// ReplacedPathHeader is the default header to set the old path to.
	ReplacedPathHeader = "X-Replaced-Path"
	typeName           = "ReplacePathFromURLRegex"
)

// Config the plugin configuration.
type Config struct {
	Regex       string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
	}
}

// ReplacePathFromURLRegex a traefik plugin.
type ReplacePathFromURLRegex struct {
	next     http.Handler
	name     string
	regexp      *regexp.Regexp
	replacement string
}

// New created a new ReplacePathFromUrlRegex plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %w", config.Regex, err)
	}
	return &ReplacePathFromURLRegex{
		next:     next,
		name:     name,
		regexp:   exp,
		replacement: strings.TrimSpace(config.Replacement),
	}, nil
}

func (r *ReplacePathFromURLRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	oldURL := rawURL(req)

	var currentPath string
	if req.URL.RawPath == "" {
		currentPath = req.URL.Path
	} else {
		currentPath = req.URL.RawPath
	}

	if r.regexp != nil && len(r.replacement) > 0 && r.regexp.MatchString(oldURL) {
		req.Header.Add(ReplacedPathHeader, currentPath)

		// req.URL.RawPath = r.regexp.ReplaceAllString(oldURL, r.replacement)
		newPath := r.regexp.ReplaceAllString(oldURL, r.replacement)
		// replace any variables that may be in there
		rewrittenPath := &bytes.Buffer{}
		if err := applyString(newPath, rewrittenPath, req); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		// as replacement can introduce escaped characters
		// Path must remain an unescaped version of RawPath
		// Doesn't handle multiple times encoded replacement (`/` => `%2F` => `%252F` => ...)
		var err error
		req.URL.Path, err = url.PathUnescape(rewrittenPath.String())
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.RequestURI = req.URL.RequestURI()
	}

	r.next.ServeHTTP(rw, req)
}


func rawURL(req *http.Request) string {
	scheme := "http"
	host := req.Host
	port := ""
	var uri string
	if req.RequestURI != "" {
		uri = req.RequestURI
	} else if req.URL.RawPath == "" {
		uri = req.URL.Path
	} else {
		uri = req.URL.RawPath
	}


	if req.TLS != nil {
		scheme = "https"
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}


func applyString(in string, out io.Writer, req *http.Request) error {
	t, err := template.New("t").Parse(in)
	if err != nil {
		return err
	}

	data := struct{ Request *http.Request }{Request: req}

	return t.Execute(out, data)
}
