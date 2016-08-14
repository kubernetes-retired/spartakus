// Package logr defines abstract interfaces for logging.  Packages can depend on
// these interfaces and callers can implement logging in whatever way is
// appropriate.
//
// This design derives from Dave Cheney's blog:
//     http://dave.cheney.net/2015/11/05/lets-talk-about-logging
package logr

// TODO: structured logging a la uber-go/zap
// TODO: consider other bits of functionality like Flush(), InfoDepth(), OutputStats

// InfoLogger represents the ability to log non-error messages.
type InfoLogger interface {
	// Info logs a non-error message.  This is behaviorally akin to fmt.Println,
	// rather than fmt.Print.
	Info(args ...interface{})

	// Infof logs a formatted non-error message.
	Infof(format string, args ...interface{})

	// Enabled test whether this InfoLogger is enabled.  For example,
	// commandline flags might be used to set the logging verbosity and disable
	// some info logs.
	Enabled() bool
}

// Logger represents the ability to log messages, both errors and not.
type Logger interface {
	// All Loggers implement InfoLogger.  Calling InfoLogger methods directly on
	// a Logger value is equivalent to calling them on a V(0) InfoLogger.  For
	// example, logger.Info() produces the same result as logger.V(0).Info.
	InfoLogger

	// Error logs a error message.  This is behaviorally akin to fmt.Println,
	// rather than fmt.Print.
	Error(args ...interface{})

	// Errorf logs a formatted error message.
	Errorf(format string, args ...interface{})

	// V returns an InfoLogger value for a specific verbosity level.
	V(level int) InfoLogger

	// NewWithPrefix returns a Logger which prefixes all messages.
	NewWithPrefix(prefix string) Logger
}
