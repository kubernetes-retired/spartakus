package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coreos/pkg/capnslog"

	"k8s.io/spartakus/pkg/volunteer"
)

var (
	VERSION = "UNKNOWN"
)

var log = capnslog.NewPackageLogger("k8s.io/spartakus/cmd", "spartakus-volunteer")

func main() {
	fs := flag.NewFlagSet("spartakus-volunteer", flag.ExitOnError)

	ver := fs.Bool("version", false, "Print version information and exit")
	logDebug := fs.Bool("log-debug", false, "Log debug-level information")

	var cfg volunteer.Config
	fs.DurationVar(&cfg.Interval, "generation-interval", volunteer.DefaultGenerationInterval, "Period of report generation attempts")
	fs.StringVar(&cfg.AccountID, "account-id", "", "Tectonic account identifier")
	fs.StringVar(&cfg.AccountSecret, "account-secret", "", "Tectonic account secret")
	fs.StringVar(&cfg.ClusterID, "cluster-id", "", "Tectonic cluster identifier")
	fs.StringVar(&cfg.CollectorScheme, "collector-scheme", "https", "Send reports to collector host using this URL scheme (http or https)")
	fs.StringVar(&cfg.CollectorHost, "collector-host", "spartakus.k8s.io", "Send generated reports to this Spartakus collector host")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing failed: %v", err)
	}

	if *ver {
		fmt.Printf("spartakus-volunteer version %s\n", VERSION)
		os.Exit(0)
	}

	capnslog.SetFormatter(capnslog.NewStringFormatter(os.Stderr))
	if *logDebug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	}

	volunteer, err := volunteer.New(cfg)
	if err != nil {
		log.Fatalf("failed building volunteer: %v", err)
	}

	volunteer.Run()
}
