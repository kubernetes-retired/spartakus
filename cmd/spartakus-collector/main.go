package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/database"
	"k8s.io/spartakus/pkg/glogr"
)

var (
	VERSION = "UNKNOWN"
)

func main() {
	fs := flag.NewFlagSet("spartakus-collector", flag.ExitOnError)

	flV := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	flVersion := fs.Bool("version", false, "Print version information and exit")
	flPort := fs.Int("port", 8080, "Port on which to listen")
	flDatabase := fs.String("database", "stdout", "Where to store records; use --print-databases for a list of options")
	flPrintDatabases := fs.Bool("print-databases", false, "Print database options and exit")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *flVersion {
		fmt.Printf("spartakus-collector version %s\n", VERSION)
		os.Exit(0)
	}

	if *flPrintDatabases {
		fmt.Printf("Example values for --database:\n")
		for _, str := range database.DatabaseOptions() {
			fmt.Println("    ", str)
		}
		os.Exit(0)
	}

	if *flPort < 1 || *flPort > 65535 {
		fmt.Fprintf(os.Stderr, "FATAL: invalid value for --port: must be between 1 and 65535")
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

	srv := &collector.APIServer{
		Log:      log,
		Port:     *flPort,
		Database: db,
		Version:  VERSION,
	}

	if err := srv.Run(); err != nil {
		log.Errorf("FATAL: %v", err)
		os.Exit(1)
	}
	log.V(0).Infof("exiting cleanly")
}
