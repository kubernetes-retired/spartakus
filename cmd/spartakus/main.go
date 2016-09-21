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
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/kubernetes-incubator/spartakus/pkg/version"
	"github.com/spf13/pflag"
	"github.com/thockin/glogr"
	"github.com/thockin/logr"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [collector | volunteer] ARGS\n", os.Args[0])
	os.Exit(1)
}

// subProgram is the "busybox" style hook for this multi-function binary.
type subProgram interface {
	AddFlags(fs *pflag.FlagSet)
	Validate() error
	Main(log logr.Logger) error
}

func main() {
	fs := pflag.NewFlagSet("spartakus", pflag.ExitOnError)

	flV := fs.Int("v", 0, "Set the logging verbosity level; higher values log more")
	flPrintVersion := fs.Bool("version", false, "Print version information and exit")

	if len(os.Args) == 1 {
		usage()
	}

	var prog subProgram
	switch os.Args[1] {
	case "collector":
		prog = collectorSubProgram{}
	case "volunteer":
		prog = volunteerSubProgram{}
	default:
		usage()
	}

	prog.AddFlags(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: flag parsing failed: %v\n", err)
		os.Exit(1)
	}

	if err := prog.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	if err := initGlog(*flV); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	if *flPrintVersion {
		fmt.Printf("spartakus-collector version %s\n", version.VERSION)
		os.Exit(0)
	}

	// Make a Logger instance.
	log, err := glogr.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v\n", err)
		os.Exit(1)
	}
	// From here on logging is available

	if err := prog.Main(log); err != nil {
		log.Errorf("exiting: %v", err)
		os.Exit(1)
	}
	log.V(0).Infof("exiting cleanly")
}

func initGlog(v int) error {
	// Force logging to stderr.
	stderrFlag := flag.Lookup("logtostderr")
	if stderrFlag == nil {
		return fmt.Errorf("can't find flag 'logtostderr'")
	}
	stderrFlag.Value.Set("true")

	// Set the V level from our own flag.
	vFlag := flag.Lookup("v")
	if vFlag == nil {
		return fmt.Errorf("can't find flag 'v'")
	}
	vFlag.Value.Set(strconv.Itoa(v))

	return nil
}
