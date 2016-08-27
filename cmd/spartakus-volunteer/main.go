package main

import (
	"fmt"
	"os"
	"time"

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

	flV := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	flVersion := fs.Bool("version", false, "Print version information and exit")
	flClusterID := fs.String("cluster-id", "", "Your cluster ID") //FIXME: rename to UID or UUID or GUID?
	flPeriod := fs.Duration("period", 24*time.Hour, "How often to send reports")
	flDatabase := fs.String("database", "http://spartakus.k8s.io", "Send reports to this database; use --print-databases for a list of options") //FIXME: default to SSL
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

	if *flClusterID == "" {
		fmt.Fprintf(os.Stderr, "FATAL: invalid value for --cluster-id: must not be empty")
		os.Exit(1)
	}
	if *flPeriod == time.Duration(0) {
		fmt.Fprintf(os.Stderr, "FATAL: invalid value for --period: must not be 0")
		os.Exit(1)
	}

	// Make a Logger instance.
	log, err := glogr.New(*flV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v", err)
		os.Exit(1)
	}
	// From here on logging is available

	db, err := database.NewDatabase(log, *flDatabase)
	if err != nil {
		log.Errorf("FATAL: failed to initialize database: %v", err)
		os.Exit(1)
	}

	volunteer, err := volunteer.New(log, *flClusterID, *flPeriod, db)
	if err != nil {
		log.Errorf("FATAL: failed building volunteer: %v", err)
		os.Exit(1)
	}

	if err := volunteer.Run(); err != nil {
		log.Errorf("FATAL: %v", err)
		os.Exit(1)
	}
	log.V(0).Infof("exiting cleanly")
}
