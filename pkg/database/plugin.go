package database

import (
	"fmt"
	"os"
	"sort"

	"github.com/thockin/logr"
)

// A database plugin is an abstract way to allocate a Database.
type plugin interface {
	// Attempt checks the dbspec and allocates a new Database instance if this
	// plugin supports it. If the dbspec is handled by this plugin, the first return
	// (is) will be true.
	Attempt(log logr.Logger, dbspec string) (is bool, db Database, err error)

	// ExampleSpec returns an example of a valid dbspec argument for Attempt.
	ExampleSpec() string
}

// All known plugins.
var plugins = map[string]plugin{}

// registerPlugin allows plugins to register themselves.
func registerPlugin(name string, plug plugin) {
	if _, found := plugins[name]; found {
		fmt.Fprintf(os.Stderr, "plugin %q was registered twice", name)
		os.Exit(1)
	}
	plugins[name] = plug
}

func DatabaseOptions() []string {
	ret := []string{}
	for name, plug := range plugins {
		ret = append(ret, fmt.Sprintf("%s: %q", name, plug.ExampleSpec()))
	}
	sort.Strings(ret)
	return ret
}
