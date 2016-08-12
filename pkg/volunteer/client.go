package volunteer

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/spartakus/pkg/report"
)

const DefaultGenerationInterval = 24 * time.Hour

type Config struct {
	ClusterID string
	Interval  time.Duration
	Collector string
}

func (cfg *Config) Valid() error {
	if cfg.ClusterID == "" {
		return fmt.Errorf("volunteer config invalid: empty cluster ID")
	}
	if cfg.Interval == time.Duration(0) {
		return fmt.Errorf("volunteer config invalid: invalid generation interval")
	}
	if cfg.Collector == "" {
		return fmt.Errorf("volunteer config invalid: empty collector")
	}
	if _, err := url.Parse(cfg.Collector); err != nil {
		return fmt.Errorf("volunteer config invalid: collector: %v", err)
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

	sender, err := newRecordSender(cfg)
	if err != nil {
		return nil, err
	}

	gen := volunteer{
		config:          cfg,
		recordSender:    sender,
		nodeLister:      kcw,
		serverVersioner: kcw,
	}

	return &gen, nil
}

type recordSender interface {
	Send(report.Record) error
}

func newRecordSender(cfg Config) (recordSender, error) {
	if cfg.Collector == "-" {
		return newStdoutRecordSender()
	} else {
		url, _ := url.Parse(cfg.Collector)
		return newHTTPRecordSender(http.DefaultClient, *url)
	}
}

type volunteer struct {
	config          Config
	recordSender    recordSender
	nodeLister      nodeLister
	serverVersioner serverVersioner
}

func (v *volunteer) Run() {
	logger.Infof("started volunteer")
	for {
		rec, err := v.generateRecord()
		if err != nil {
			logger.Errorf("failed generating report: %v", err)
			continue
		}

		if err = v.send(rec); err != nil {
			logger.Errorf("failed sending report: %v", err)
			continue
		}

		logger.Infof("report successfully sent to collector")
		logger.Infof("next attempt in %v", v.config.Interval)
		<-time.After(v.config.Interval)
	}
	return
}

func (v *volunteer) generateRecord() (report.Record, error) {
	svrVer, err := v.serverVersioner.ServerVersion()
	if err != nil {
		return report.Record{}, err
	}

	nodes, err := v.nodeLister.List()
	if err != nil {
		return report.Record{}, err
	}

	rec := report.Record{
		ID:            v.config.ClusterID,
		MasterVersion: &svrVer,
		Nodes:         nodes,
	}

	return rec, nil
}

func (v *volunteer) send(rec report.Record) error {
	return v.recordSender.Send(rec)
}
