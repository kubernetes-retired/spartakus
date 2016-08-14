// Package glogr implements k8s.io/spartakus/pkg/logr.Logger in terms of
// github.com/golang/glog.
package glogr

// TODO: a clean way for apps to set the --vmodule flag
// TODO: a clean way for apps to set the --log_backtrace_at flag

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/spartakus/pkg/logr"
)

// New returns a logr.Logger which is implemented by glog.

// Because glog only offers global logging functions (as opposed to logger
// objects), this will set some global parameters.  For example, if you call
// this more than once with different v arguments, the global verbosty level
// will change for all instances.
//
// Because glog does not offer an option to log to an arbitrary output, this
// forces logging to go to stderr.
func New(v int) (logr.Logger, error) {
	// Force logging to stderr.
	stderrFlag := flag.Lookup("logtostderr")
	if stderrFlag == nil {
		return nil, fmt.Errorf("can't find flag 'logtostderr'")
	}
	stderrFlag.Value.Set("true")

	// Set the V level.
	vFlag := flag.Lookup("v")
	if vFlag == nil {
		return nil, fmt.Errorf("can't find flag 'v'")
	}
	vFlag.Value.Set(strconv.Itoa(v))

	return glogger{
		level:  0,
		prefix: "",
	}, nil
}

type glogger struct {
	level  int
	prefix string
}

func prepend(prefix interface{}, args []interface{}) []interface{} {
	return append([]interface{}{prefix}, args...)
}

func (l glogger) Info(args ...interface{}) {
	glog.InfoDepth(1, prepend(l.prefix, args)...)
}

func (l glogger) Infof(format string, args ...interface{}) {
	glog.Infof("%s"+format, prepend(l.prefix, args)...)
}

func (l glogger) Enabled() bool {
	return bool(glog.V(glog.Level(l.level)))
}

func (l glogger) Error(args ...interface{}) {
	glog.ErrorDepth(1, prepend(l.prefix, args)...)
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
