// Package glogr implements k8s.io/spartakus/pkg/logr.Logger in terms of
// github.com/golang/glog.
package glogr

// TODO: a clean way for apps to set the --v flag
// TODO: a clean way for apps to set the --vmodule flag
// TODO: a clean way for apps to set the --log_backtrace_at flag

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"k8s.io/spartakus/pkg/logr"
)

func New() (logr.Logger, error) {
	// Force logging to stderr.
	toStderr := flag.Lookup("logtostderr")
	if toStderr == nil {
		return nil, fmt.Errorf("can't find flag 'logtostderr'")
	}
	toStderr.Value.Set("true")

	return glogger{
		level:  0,
		prefix: "",
	}, nil
}

func NewMustSucceed() logr.Logger {
	logger, err := New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ABORT: failed to initialize glogr: %v", err)
		os.Exit(7)
	}
	return logger
}

type glogger struct {
	level  int
	prefix string
}

func prepend(prefix interface{}, args []interface{}) []interface{} {
	return append([]interface{}{prefix}, args...)
}

func (l glogger) Info(args ...interface{}) {
	glog.Infoln(prepend(l.prefix, args)...)
}

func (l glogger) Infof(format string, args ...interface{}) {
	glog.Infof("%s"+format, prepend(l.prefix, args)...)
}

func (l glogger) Enabled() bool {
	return bool(glog.V(glog.Level(l.level)))
}

func (l glogger) Error(args ...interface{}) {
	glog.Errorln(prepend(l.prefix, args)...)
}

func (l glogger) Errorf(format string, args ...interface{}) {
	glog.Errorf("%s"+format, prepend(l.prefix, args)...)
}

func (l glogger) V(level int) logr.InfoLogger {
	return glogger{
		level:  level,
		prefix: l.prefix,
	}
}

func (l glogger) NewWithPrefix(prefix string) logr.Logger {
	return glogger{
		level:  l.level,
		prefix: prefix,
	}
}

var _ logr.Logger = glogger{}
var _ logr.InfoLogger = glogger{}
