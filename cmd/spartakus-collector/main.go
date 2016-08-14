package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coreos/pkg/capnslog"

	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/report"
)

var (
	VERSION = "UNKNOWN"
)

var log = capnslog.NewPackageLogger("k8s.io/spartakus/cmd", "spartakus-collector")

func main() {
	fs := flag.NewFlagSet("spartakus-collector", flag.ExitOnError)

	ver := fs.Bool("version", false, "Print version information and exit")
	logDebug := fs.Bool("log-debug", false, "Log debug-level information")
	//FIXME
	//logQueries := fs.Bool("log-queries", false, "Log all database queries")

	port := fs.Int("port", 8080, "Port on which to listen")

	//FIXME: BigQuery config
	//var cfg collector.DBConfig
	//fs.StringVar(&cfg.DSN, "db-url", "", "DSN-formatted database connection string")
	//fs.IntVar(&cfg.MaxOpenConnections, "db-max-open-conns", 0, "Maximum number of open connections to the database")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing failed: %v", err)
	}

	if *ver {
		fmt.Printf("spartakus-collector version %s\n", VERSION)
		os.Exit(0)
	}

	capnslog.SetFormatter(capnslog.NewStringFormatter(os.Stderr))
	if *logDebug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	}

	if *port < 1 || *port > 65535 {
		log.Fatalf("invalid value for --port: must be between 1 and 65535")
	}

	// FIXME: BigQuery
	//conn, err := collector.NewDBConnection(cfg)
	//if err != nil {
	//log.Fatalf("failed building DB connection: %v", err)
	//}

	srv := &collector.APIServer{
		Port: *port,
		//FIXME: BigQuery
		//Database: collector.NewDBRecordRepo(conn),
		Database: nullDatabase{},
		Version:  VERSION,
		//LogQueries: *logQueries, //FIXME: pass logger to apiserver
	}
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
	log.Infof("server exiting cleanly")
}

type nullDatabase struct{}

func (ndb nullDatabase) Store(_ report.Record) error {
	return nil
}
