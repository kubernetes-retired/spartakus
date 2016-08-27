package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/database"
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
	flVersion := fs.Bool("version", false, "Print version information and exit")

	var cfg volunteer.Config
	fs.DurationVar(&cfg.Interval, "generation-interval", volunteer.DefaultGenerationInterval, "Period of report generation attempts")
	//FIXME: rename to UID or UUID or GUID?
	fs.StringVar(&cfg.ClusterID, "cluster-id", "", "Your cluster GUID")

	//FIXME: default to SSL
	flDatabase := fs.String("database", "http://spartakus.k8s.io", "Send reports to this database; use --print-databases for a list of options")
	flPrintDatabases := fs.Bool("print-databases", false, "Print database options and exit")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *flVersion {
		fmt.Printf("spartakus-volunteer version %s\n", VERSION)
		os.Exit(0)
	}

	if *flPrintDatabases {
		fmt.Printf("Example values for --database:\n")
		for _, str := range database.DatabaseOptions() {
			fmt.Println("    ", str)
		}
		os.Exit(0)
	}

	if logger, err := glogr.New(*vFlag); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v", err)
		os.Exit(1)
	} else {
		log = logger // set the global
	}
	// From here on logging is available

	db, err := database.NewDatabase(log, *flDatabase)
	if err != nil {
		log.Errorf("FATAL: failed to initialize database: %v", err)
		os.Exit(1)
	}
	cfg.Database = db

	volunteer, err := volunteer.New(cfg, log)
	if err != nil {
		log.Errorf("FATAL: failed building volunteer: %v", err)
		os.Exit(1)
	}

	volunteer.Run()
}
