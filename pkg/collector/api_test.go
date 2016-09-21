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

package collector

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	logrtest "github.com/thockin/logr/testing"
)

var (
	httpHeaderJSONContentType = http.Header{"Content-Type": []string{"application/json"}}
	minimumRecordJSON         = `{}`
)

// HandlerRoundTripper implements the net/http.RoundTripper using
// an in-memory net/http.Handler, skipping any network calls.
type HandlerRoundTripper struct {
	Handler http.Handler
}

func (rt *HandlerRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	rt.Handler.ServeHTTP(w, r)

	resp := http.Response{
		StatusCode: w.Code,
		Header:     w.Header(),
		Body:       ioutil.NopCloser(w.Body),
	}

	return &resp, nil
}

type memDatabase struct {
	Records []report.Record
}

func (d *memDatabase) Store(r report.Record) error {
	d.Records = append(d.Records, r)
	return nil
}

type testServer struct {
	Database database.Database
	srv      *APIServer
}

func (ts *testServer) URL(p string) string {
	u := url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   p,
	}
	return u.String()
}

func (ts *testServer) Server(t *testing.T) *APIServer {
	if ts.srv == nil {
		ts.srv = &APIServer{
			Log:      logrtest.TestLogger{t},
			Database: ts.Database,
		}
	}
	return ts.srv
}

func (ts *testServer) HTTPClient(t *testing.T) *http.Client {
	return &http.Client{
		Transport: &HandlerRoundTripper{
			Handler: ts.Server(t).newHandler(),
		},
	}
}

type errDatabase struct {
	err error
}

func (d *errDatabase) Store(report.Record) error {
	return d.err
}

func TestRecordResourceStoreBadContentType(t *testing.T) {
	db := &memDatabase{}
	srv := &testServer{Database: db}
	cli := srv.HTTPClient(t)

	req, err := http.NewRequest("POST", srv.URL(CollectorEndpoint), strings.NewReader(minimumRecordJSON))
	if err != nil {
		t.Fatalf("unable to create HTTP request: %v", err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		t.Fatalf("unable to get HTTP response: %v", err)
	}

	wantStatusCode := http.StatusUnsupportedMediaType
	if wantStatusCode != resp.StatusCode {
		t.Fatalf("incorrect status code: want=%d got=%d", wantStatusCode, resp.StatusCode)
	}
}

func TestRecordResourceStoreInternalError(t *testing.T) {
	db := &errDatabase{err: errors.New("fail!")}
	srv := &testServer{Database: db}
	cli := srv.HTTPClient(t)

	req, err := http.NewRequest("POST", srv.URL(CollectorEndpoint), strings.NewReader(minimumRecordJSON))
	if err != nil {
		t.Fatalf("unable to create HTTP request: %v", err)
	}

	req.Header = httpHeaderJSONContentType

	resp, err := cli.Do(req)
	if err != nil {
		t.Fatalf("unable to get HTTP response: %v", err)
	}

	wantStatusCode := http.StatusInternalServerError
	if wantStatusCode != resp.StatusCode {
		t.Fatalf("incorrect status code: want=%d got=%d", wantStatusCode, resp.StatusCode)
	}
}

func TestRecordResourceStoreBadRecord(t *testing.T) {
	tests := []string{
		// invalid JSON
		`{`,
	}
	for i, tt := range tests {
		db := &memDatabase{}
		srv := &testServer{Database: db}
		cli := srv.HTTPClient(t)

		req, err := http.NewRequest("POST", srv.URL(CollectorEndpoint), strings.NewReader(tt))
		if err != nil {
			t.Fatalf("case %d: unable to create HTTP request: %v", i, err)
		}

		req.Header = httpHeaderJSONContentType

		resp, err := cli.Do(req)
		if err != nil {
			t.Fatalf("case %d: unable to get HTTP response: %v", i, err)
		}

		wantStatusCode := http.StatusBadRequest
		if wantStatusCode != resp.StatusCode {
			t.Fatalf("case %d: incorrect status code: want=%d got=%d", i, wantStatusCode, resp.StatusCode)
		}
	}
}

func TestRecordResoureStoreSuccess(t *testing.T) {
	db := &memDatabase{}
	srv := &testServer{Database: db}
	cli := srv.HTTPClient(t)

	req, err := http.NewRequest("POST", srv.URL(CollectorEndpoint), strings.NewReader(minimumRecordJSON))
	if err != nil {
		t.Fatalf("unable to create HTTP request: %v", err)
	}

	req.Header = httpHeaderJSONContentType

	resp, err := cli.Do(req)
	if err != nil {
		t.Fatalf("unable to get HTTP response: %v", err)
	}

	wantStatusCode := http.StatusNoContent
	if wantStatusCode != resp.StatusCode {
		t.Fatalf("incorrect status code: want=%d got=%d", wantStatusCode, resp.StatusCode)
	}
}
