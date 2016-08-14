package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/spartakus/pkg/glogr"
	"k8s.io/spartakus/pkg/volunteer"
)

var (
	VERSION = "UNKNOWN"
)

var log = glogr.NewMustSucceed()

func main() {
	fs := flag.NewFlagSet("spartakus-volunteer", flag.ExitOnError)

	ver := fs.Bool("version", false, "Print version information and exit")

	var cfg volunteer.Config
	fs.DurationVar(&cfg.Interval, "generation-interval", volunteer.DefaultGenerationInterval, "Period of report generation attempts")
	//FIXME: rename to UID or UUID or GUID?
	fs.StringVar(&cfg.ClusterID, "cluster-id", "", "Your cluster GUID")
	//FIXME: need SSL
	fs.StringVar(&cfg.Collector, "collector", "http://spartakus.k8s.io", "Send generated reports to this Spartakus collector (use - for stdout)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Errorf("FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *ver {
		fmt.Printf("spartakus-volunteer version %s\n", VERSION)
		os.Exit(0)
	}

	volunteer, err := volunteer.New(cfg, log)
	if err != nil {
		log.Errorf("FATAL: failed building volunteer: %v", err)
		os.Exit(1)
	}

	volunteer.Run()
}
