package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/glogr"
	"k8s.io/spartakus/pkg/logr"
)

var (
	VERSION = "UNKNOWN"
)

func main() {
	fs := flag.NewFlagSet("spartakus-collector", flag.ExitOnError)

	flV := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	flVersion := fs.Bool("version", false, "Print version information and exit")
	flPort := fs.Int("port", 8080, "Port on which to listen")
	flDatabase := fs.String("database", "stdout", "Where to store records; may be be 'stdout' or 'bigquery://project.dataset.table'")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v", err)
		os.Exit(1)
	}

	if *flPort < 1 || *flPort > 65535 {
		fmt.Fprintf(os.Stderr, "FATAL: invalid value for --port: must be between 1 and 65535")
		os.Exit(1)
	}

	if *flVersion {
		fmt.Printf("spartakus-collector version %s\n", VERSION)
		os.Exit(0)
	}

	// Make a Logger instance.
	log, err := glogr.New(*flV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v", err)
		os.Exit(1)
	}
	// From here on logging is available

	db, err := newDatabase(log, *flDatabase)
	if err != nil {
		log.Errorf("failed to create database: %v", err)
		os.Exit(1)
	}

	srv := &collector.APIServer{
		Log:      log,
		Port:     *flPort,
		Database: db,
		Version:  VERSION,
	}

	if err := srv.Run(); err != nil {
		log.Errorf("FATAL: server error: %v", err)
		os.Exit(1)
	}
	log.V(0).Infof("server exiting cleanly")
}

func newDatabase(log logr.Logger, dbspec string) (collector.Database, error) {
	if dbspec == "stdout" {
		return stdoutDatabase{}, nil
	}
	// Check if the dbspec is a bigquery spec, and parse it, if so.
	if is, project, dataset, table, err := parseBigquerySpec(dbspec); is {
		if err != nil {
			return nil, err
		}
		return newBigqueryDatabase(log, project, dataset, table)
	}
	return nil, fmt.Errorf("invalid database specification: %q", dbspec)
}

var bqre = regexp.MustCompile(`^bigquery://([^.]+)\.([^.]+)\.([^.]+)$`)

// Parse a bigquery spec, formatted as `bigquery://project.dataset.table`.  If
// the argument appears to be a bigquery spec, the first return (bool) will be
// true. The 3 string returns are project, dataset, and table respectively.
// This will only return an error if it believes the argument is a biquery spec,
// but it can't parse properly.
func parseBigquerySpec(dbspec string) (bool, string, string, string, error) {
	if !strings.HasPrefix(dbspec, "bigquery://") {
		return false, "", "", "", nil
	}
	subs := bqre.FindStringSubmatch(dbspec)
	if len(subs) != 3 {
		return true, "", "", "", fmt.Errorf("invalid bigquery spec: %q", dbspec)
	}
	return true, subs[0], subs[1], subs[2], nil
}
