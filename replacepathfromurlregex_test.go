package replacepathfromurlregex_test

import (
	"context"
	"github.com/rocdove/replacepathfromurlregex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplacePathFromUrlRegex(t *testing.T) {
	cfg := replacepathfromurlregex.CreateConfig()
	cfg.Regex = "^https?://([a-z]+).demo.localhost(:[0-9]+)?/c/prefix/(.*?)$"
	cfg.Replacement = "/${1}/${3}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := replacepathfromurlregex.New(ctx, next, cfg, "replacePathFromURLRegex-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://a.demo.localhost:80/c/prefix/uri", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertPath(t, req, "/a/uri")

}

func assertPath(t *testing.T, req *http.Request, expected string) {
	t.Helper()

	var currentPath string
	if req.URL.RawPath == "" {
		currentPath = req.URL.Path
	} else {
		currentPath = req.URL.RawPath
	}

	if currentPath != expected {
		t.Errorf("invalid RawPath value: %s, expected: %s", currentPath, expected)
	}
}
