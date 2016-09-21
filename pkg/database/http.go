/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/thockin/logr"
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
