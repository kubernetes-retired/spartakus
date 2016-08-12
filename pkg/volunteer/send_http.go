package volunteer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/spartakus/pkg/report"
)

var (
	statsEndpoint = "/api/v1/stats"
)

// newHTTPRecordSender returns a RecordSender that interacts with a
// collector API using the provided HTTP client.
func newHTTPRecordSender(c *http.Client, u url.URL) (*httpRecordSender, error) {
	p, err := u.Parse(statsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("volunteer: unable to prepare API URL: %v", err)
	}

	r := &httpRecordSender{
		statsURL: p.String(),
		client:   c,
	}

	return r, nil
}

type httpRecordSender struct {
	statsURL string
	client   *http.Client
}

func (h *httpRecordSender) Send(r report.Record) error {
	body, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("volunteer: unable to encode HTTP request body: %v", err)
	}
	req, err := http.NewRequest("POST", h.statsURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("volunteer: unable to prepare HTTP request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("volunteer: HTTP request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("volunteer: received unexpected HTTP response code %d", res.StatusCode)
	}
	return nil
}
