package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/spartakus/pkg/glogr"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/version"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [collector | volunteer] ARGS\n", os.Args[0])
	os.Exit(1)
}

// subProgram is the "busybox" style hook for this multi-function binary.
type subProgram interface {
	AddFlags(fs *flag.FlagSet)
	Validate() error
	Main(log logr.Logger) error
}

func main() {
	fs := flag.NewFlagSet("spartakus", flag.ExitOnError)

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

	if *flPrintVersion {
		fmt.Printf("spartakus-collector version %s\n", version.VERSION)
		os.Exit(0)
	}

	// Make a Logger instance.
	log, err := glogr.New(*flV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize glogr: %v\n", err)
		os.Exit(1)
	}
	// From here on logging is available

	if err := prog.Main(log); err != nil {
		log.Errorf("exiting: %v", err)
	}
	log.V(0).Infof("exiting cleanly")
}
