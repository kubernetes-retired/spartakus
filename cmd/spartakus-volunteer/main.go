package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/glogr"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/volunteer"
)

var (
	VERSION = "UNKNOWN"
)

var log logr.Logger

func main() {
	fs := flag.NewFlagSet("spartakus-volunteer", flag.ExitOnError)

	vFlag := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	versionFlag := fs.Bool("version", false, "Print version information and exit")

	var cfg volunteer.Config
	fs.DurationVar(&cfg.Interval, "generation-interval", volunteer.DefaultGenerationInterval, "Period of report generation attempts")
	//FIXME: rename to UID or UUID or GUID?
	fs.StringVar(&cfg.ClusterID, "cluster-id", "", "Your cluster GUID")
	//FIXME: need SSL
	fs.StringVar(&cfg.Collector, "collector", "http://spartakus.k8s.io", "Send generated reports to this Spartakus collector (use - for stdout)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *versionFlag {
		fmt.Printf("spartakus-volunteer version %s\n", VERSION)
		os.Exit(0)
	}

	if logger, err := glogr.New(*vFlag); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v", err)
		os.Exit(1)
	} else {
		log = logger // set the global
	}
	// From here on logging is available

	volunteer, err := volunteer.New(cfg, log)
	if err != nil {
		log.Errorf("FATAL: failed building volunteer: %v", err)
		os.Exit(1)
	}

	volunteer.Run()
}
