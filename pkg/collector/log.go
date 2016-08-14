package collector

import (
	"k8s.io/spartakus/pkg/logr"
)

type logWriter struct {
	log logr.Logger
	lvl int
}

func (l logWriter) Write(d []byte) (int, error) {
	l.log.V(l.lvl).Infof(string(d))
	return len(d), nil
}
