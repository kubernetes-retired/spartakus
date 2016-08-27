package volunteer

import (
	"fmt"
	"time"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/spartakus/pkg/database"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/report"
)

const DefaultGenerationInterval = 24 * time.Hour

type Config struct {
	ClusterID string
	Interval  time.Duration
	Database  database.Database
}

func (cfg *Config) Valid() error {
	if cfg.ClusterID == "" {
		return fmt.Errorf("volunteer config invalid: empty cluster ID")
	}
	if cfg.Interval == time.Duration(0) {
		return fmt.Errorf("volunteer config invalid: invalid generation interval")
	}
	if cfg.Database == nil {
		return fmt.Errorf("volunteer config invalid: no database")
	}
	return nil
}

func New(cfg Config, log logr.Logger) (*volunteer, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	kc, err := kclient.NewInCluster()
	if err != nil {
		return nil, err
	}
	kcw := &kubernetesClientWrapper{client: kc}

	gen := volunteer{
		config:          cfg,
		log:             log,
		nodeLister:      kcw,
		serverVersioner: kcw,
	}

	return &gen, nil
}

type volunteer struct {
	config          Config
	log             logr.Logger
	nodeLister      nodeLister
	serverVersioner serverVersioner
}

func (v *volunteer) Run() {
	v.log.V(0).Infof("started volunteer")
	for {
		rec, err := v.generateRecord()
		if err != nil {
			v.log.Errorf("failed generating report: %v", err)
			continue
		}

		if err = v.send(rec); err != nil {
			v.log.Errorf("failed sending report: %v", err)
			continue
		}

		v.log.V(0).Infof("report successfully sent to collector")
		v.log.V(0).Infof("next attempt in %v", v.config.Interval)
		<-time.After(v.config.Interval)
	}
	return
}

func (v *volunteer) generateRecord() (report.Record, error) {
	svrVer, err := v.serverVersioner.ServerVersion()
	if err != nil {
		return report.Record{}, err
	}

	nodes, err := v.nodeLister.ListNodes()
	if err != nil {
		return report.Record{}, err
	}

	rec := report.Record{
		Version:       "abc123", //FIXME: from linker
		ClusterID:     v.config.ClusterID,
		MasterVersion: &svrVer,
		Nodes:         nodes,
	}

	return rec, nil
}

func (v *volunteer) send(rec report.Record) error {
	return v.config.Database.Store(rec)
}
