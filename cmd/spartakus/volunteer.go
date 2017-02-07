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
	"time"

	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/volunteer"
	"github.com/spf13/pflag"
	"github.com/thockin/logr"
)

var volunteerConfig = struct {
	clusterID      string
	period         time.Duration
	database       string
	printDatabases bool
	extensionsPath string
}{}

type volunteerSubProgram struct{}

func (_ volunteerSubProgram) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&volunteerConfig.clusterID, "cluster-id", "", "Your cluster ID")
	fs.DurationVar(&volunteerConfig.period, "period", 24*time.Hour, "How often to send reports; set to 0 for one-shot mode")
	fs.StringVar(&volunteerConfig.database, "database",
		"https://spartakus.k8s.io", "Send reports to this database; use --print-databases for a list of options")
	fs.BoolVar(&volunteerConfig.printDatabases, "print-databases", false, "Print database options and exit")
	fs.StringVar(&volunteerConfig.extensionsPath, "extensions", "", "Path to a file of additional metrics to report; leave unset to report no additional metrics")
}

func (_ volunteerSubProgram) Validate() error {
	if volunteerConfig.clusterID == "" {
		return fmt.Errorf("invalid value for --cluster-id: must not be empty")
	}
	return nil
}

func (_ volunteerSubProgram) Main(log logr.Logger) error {
	if volunteerConfig.printDatabases {
		fmt.Printf("Example values for --database:\n")
		for _, str := range database.DatabaseOptions() {
			fmt.Println("    ", str)
		}
		os.Exit(0)
	}

	db, err := database.NewDatabase(log, volunteerConfig.database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	volunteer, err := volunteer.New(log, volunteerConfig.clusterID, volunteerConfig.period, db, volunteerConfig.extensionsPath)
	if err != nil {
		return fmt.Errorf("failed initializing volunteer: %v", err)
	}

	if err := volunteer.Run(); err != nil {
		return err
	}
	return nil
}
