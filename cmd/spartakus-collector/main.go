package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/glogr"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/report"
)

var (
	VERSION = "UNKNOWN"
)

// Global logger instance.
var log logr.Logger

func main() {
	fs := flag.NewFlagSet("spartakus-collector", flag.ExitOnError)

	vFlag := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	versionFlag := fs.Bool("version", false, "Print version information and exit")
	//FIXME
	//logQueries := fs.Bool("log-queries", false, "Log all database queries")

	port := fs.Int("port", 8080, "Port on which to listen")

	//FIXME: BigQuery config
	//var cfg collector.DBConfig
	//fs.StringVar(&cfg.DSN, "db-url", "", "DSN-formatted database connection string")
	//fs.IntVar(&cfg.MaxOpenConnections, "db-max-open-conns", 0, "Maximum number of open connections to the database")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *port < 1 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "FATAL: invalid value for --port: must be between 1 and 65535")
		os.Exit(1)
	}

	if *versionFlag {
		fmt.Printf("spartakus-collector version %s\n", VERSION)
		os.Exit(0)
	}

	if logger, err := glogr.New(*vFlag); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v", err)
		os.Exit(1)
	} else {
		log = logger // set the global
	}
	// From here on logging is available

	// FIXME: BigQuery
	//conn, err := collector.NewDBConnection(cfg)
	//if err != nil {
	//log.Fatalf("failed building DB connection: %v", err)
	//}

	// FIXME: wrap in a New()
	srv := &collector.APIServer{
		Log:  log,
		Port: *port,
		//FIXME: BigQuery
		//Database: collector.NewDBRecordRepo(conn),
		Database: nullDatabase{},
		Version:  VERSION,
		//LogQueries: *logQueries, //FIXME: pass logger to apiserver
	}
	if err := srv.Run(); err != nil {
		log.Errorf("FATAL: server error: %v", err)
		os.Exit(1)
	}
	log.V(0).Infof("server exiting cleanly")
}

type nullDatabase struct{}

func (ndb nullDatabase) Store(_ report.Record) error {
	return nil
}
