/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kubernetes-incubator/spartakus/pkg/collector"
	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/spf13/pflag"
	"github.com/thockin/logr"
)

var collectorConfig = struct {
	port           int
	database       string
	printDatabases bool
}{}

type collectorSubProgram struct{}

func (_ collectorSubProgram) AddFlags(fs *pflag.FlagSet) {
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

	go handleSignals()

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

func handleSignals() {
	c := make(chan os.Signal, 1)
	// Trap and ignore SIGTERM.
	signal.Notify(c, syscall.SIGTERM)
	for {
		s := <-c
		fmt.Printf("Got signal: %v\n", s)
	}
}
