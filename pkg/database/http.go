package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thockin/logr"
	"k8s.io/spartakus/pkg/report"
)

func init() {
	registerPlugin("http", httpPlugin{})
}

// This plugin POSTS a JSON-encoded report.Record to a URL at /api/v1 path.
type httpPlugin struct{}

func (plug httpPlugin) Attempt(log logr.Logger, dbspec string) (bool, Database, error) {
	if !strings.HasPrefix(dbspec, "http://") && !strings.HasPrefix(dbspec, "https://") {
		return false, nil, nil
	}

	url, err := url.Parse(dbspec)
	if err != nil {
		return true, nil, fmt.Errorf("invalid http spec: %q: %v", dbspec, err)
	}
	db, err := newHTTPDatabase(newHTTPClient(), *url)
	return true, db, err
}

func (plug httpPlugin) ExampleSpec() string {
	return "http://spartakus.example.com"
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{
		Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableCompression:  true,
	}
	return &http.Client{Transport: tr}
}

var urlPath = "/api/v1"

func newHTTPDatabase(c *http.Client, u url.URL) (Database, error) {
	p, err := u.Parse(urlPath)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare API URL: %v", err)
	}

	db := &httpDatabase{
		client: c,
		url:    p.String(),
	}

	return db, nil
}

type httpDatabase struct {
	url    string
	client *http.Client
}

func (h *httpDatabase) Store(r report.Record) error {
	body, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("unable to encode HTTP request body: %v", err)
	}

	req, err := http.NewRequest("POST", h.url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to prepare HTTP request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("received unexpected HTTP response code %d", res.StatusCode)
	}

	return nil
}
