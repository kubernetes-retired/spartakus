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
	DefaultGenerationInterval = time.Hour

	StatsMetadataFieldGeneratedAt = "generated_at"
)

type TectonicPayload struct {
	KubernetesNodes []node `json:"kubernetesNodes"`
	ClusterID       string `json:"clusterID"`
}

type Config struct {
	AccountID       string
	AccountSecret   string
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
	if cfg.AccountID == "" {
		return errors.New("volunteer config invalid: empty account ID")
	}
	if cfg.AccountSecret == "" {
		return errors.New("volunteer config invalid: empty account secret")
	}
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
	log.Infof("started volunteer")
	for {
		log.Infof("next attempt in %v", v.config.Interval)
		<-time.After(v.config.Interval)

		p, err := v.payload()
		if err != nil {
			log.Errorf("failed generating report: %v", err)
			continue
		}

		if err = v.send(p); err != nil {
			log.Errorf("failed sending report: %v", err)
			continue
		}

		log.Infof("report successfully sent to collector")
	}
	return
}

func (v *volunteer) payload() (TectonicPayload, error) {
	nodes, err := v.nodeLister.List()
	if err != nil {
		return TectonicPayload{}, err
	}

	//TODO(bcwaldon): add license and cluster version info
	p := TectonicPayload{
		ClusterID:       v.config.ClusterID,
		KubernetesNodes: nodes,
	}

	return p, nil
}

func (v *volunteer) send(p TectonicPayload) error {
	m := map[string]string{
		StatsMetadataFieldGeneratedAt: strconv.FormatInt(time.Now().Unix(), 10),
	}

	r := report.Record{
		AccountID:     v.config.AccountID,
		AccountSecret: v.config.AccountSecret,
		Metadata:      m,
		Payload:       p,
	}

	return v.recordRepo.Store(r)
}
