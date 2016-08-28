package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/database"
	"k8s.io/spartakus/pkg/logr"
)

var collectorConfig = struct {
	port           int
	database       string
	printDatabases bool
}{}

type collectorSubProgram struct{}

func (_ collectorSubProgram) AddFlags(fs *flag.FlagSet) {
	fs.IntVar(&collectorConfig.port, "port", 8080, "Port on which to listen")
	fs.StringVar(&collectorConfig.database, "database", "stdout", "Where to store records; use --print-databases for a list of options")
	fs.BoolVar(&collectorConfig.printDatabases, "print-databases", false, "Print database options and exit")
}

func (_ collectorSubProgram) Validate() error {
	if collectorConfig.port < 1 || collectorConfig.port > 65535 {
		return fmt.Errorf("invalid value for --port: must be between 1 and 65535")
	}
	return nil
}

func (_ collectorSubProgram) Main(log logr.Logger) error {
	if collectorConfig.printDatabases {
		fmt.Printf("Example values for --database:\n")
		for _, str := range database.DatabaseOptions() {
			fmt.Println("    ", str)
		}
		os.Exit(0)
	}

	db, err := database.NewDatabase(log, collectorConfig.database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	srv := &collector.APIServer{
		Log:      log,
		Port:     collectorConfig.port,
		Database: db,
	}

	if err := srv.Run(); err != nil {
		return err
	}
	return nil
}
