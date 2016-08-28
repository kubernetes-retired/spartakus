package volunteer

import (
	"fmt"
	"time"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/spartakus/pkg/database"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/report"
	"k8s.io/spartakus/pkg/version"
)

func New(log logr.Logger, clusterID string, period time.Duration, db database.Database) (*volunteer, error) {
	kc, err := kclient.NewInCluster()
	if err != nil {
		return nil, err
	}
	kcw := &kubernetesClientWrapper{client: kc}

	gen := volunteer{
		log:             log,
		clusterID:       clusterID,
		period:          period,
		database:        db,
		nodeLister:      kcw,
		serverVersioner: kcw,
	}

	return &gen, nil
}

type volunteer struct {
	clusterID       string
	period          time.Duration
	database        database.Database
	log             logr.Logger
	nodeLister      nodeLister
	serverVersioner serverVersioner
}

func (v *volunteer) Run() error {
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
		v.log.V(0).Infof("next attempt in %v", v.period)
		<-time.After(v.period)
	}
	return fmt.Errorf("unexpected termination")
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
		Version:       version.VERSION,
		ClusterID:     v.clusterID,
		MasterVersion: &svrVer,
		Nodes:         nodes,
	}

	return rec, nil
}

func (v *volunteer) send(rec report.Record) error {
	return v.database.Store(rec)
}
