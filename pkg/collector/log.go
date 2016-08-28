package collector

import (
	"github.com/thockin/logr"
)

type logWriter struct {
	log logr.Logger
	lvl int
}

func (l logWriter) Write(d []byte) (int, error) {
	l.log.V(l.lvl).Infof(string(d))
	return len(d), nil
}
