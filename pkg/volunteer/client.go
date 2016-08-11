package volunteer

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/spartakus/pkg/report"
)

var (
	DefaultGenerationInterval = 24 * time.Hour

	MetadataFieldTimestamp = "timestamp"
)

type Payload struct {
	Nodes     []node `json:"nodes"`
	ClusterID string `json:"clusterID"`
}

type Config struct {
	ClusterID       string
	Interval        time.Duration
	CollectorScheme string
	CollectorHost   string
}

func (cfg *Config) CollectorURL() url.URL {
	return url.URL{Scheme: cfg.CollectorScheme, Host: cfg.CollectorHost}
}

func (cfg *Config) CollectorHTTPClient() *http.Client {
	return http.DefaultClient
}

func (cfg *Config) Valid() error {
	if cfg.ClusterID == "" {
		return errors.New("volunteer config invalid: empty cluster ID")
	}
	if cfg.Interval == time.Duration(0) {
		return errors.New("volunteer config invalid: invalid generation interval")
	}
	if cfg.CollectorScheme != "http" && cfg.CollectorScheme != "https" {
		return errors.New("volunteer config invalid: invalid collector scheme, must be http/https")
	}
	if cfg.CollectorHost == "" {
		return errors.New("volunteer config invalid: empty collector host")
	}
	return nil
}

func New(cfg Config) (*volunteer, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	kc, err := kclient.NewInCluster()
	if err != nil {
		return nil, err
	}
	kcw := &kubernetesClientWrapper{client: kc}

	rRepo, err := NewHTTPRecordRepo(cfg.CollectorHTTPClient(), cfg.CollectorURL())
	if err != nil {
		return nil, err
	}

	gen := volunteer{
		config:     cfg,
		recordRepo: rRepo,
		nodeLister: kcw,
	}

	return &gen, nil
}

type volunteer struct {
	config     Config
	nodeLister nodeLister
	recordRepo report.RecordRepo
}

func (v *volunteer) Run() {
	logger.Infof("started volunteer")
	for {
		logger.Infof("next attempt in %v", v.config.Interval)
		<-time.After(v.config.Interval)

		p, err := v.payload()
		if err != nil {
			logger.Errorf("failed generating report: %v", err)
			continue
		}

		if err = v.send(p); err != nil {
			logger.Errorf("failed sending report: %v", err)
			continue
		}

		logger.Infof("report successfully sent to collector")
	}
	return
}

func (v *volunteer) payload() (Payload, error) {
	nodes, err := v.nodeLister.List()
	if err != nil {
		return Payload{}, err
	}

	//TODO(bcwaldon): add license and cluster version info
	p := Payload{
		ClusterID: v.config.ClusterID,
		Nodes:     nodes,
	}

	return p, nil
}

func (v *volunteer) send(p Payload) error {
	m := map[string]string{
		MetadataFieldTimestamp: strconv.FormatInt(time.Now().Unix(), 10),
	}

	r := report.Record{
		Metadata: m,
		Payload:  p,
	}

	return v.recordRepo.Store(r)
}
