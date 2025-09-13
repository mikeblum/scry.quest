// Package param provides CLI command and flag constants
package param

// Flag represents CLI flag types
type Flag string

const (
	// FlagLogLevel sets the logging level
	FlagLogLevel Flag = "log-level"
	// FlagLogFormat sets the log output format
	FlagLogFormat Flag = "log-format"
	// FlagJSON enables JSON output format
	FlagJSON Flag = "json"
	// FlagText enables text output format
	FlagText Flag = "text"
	// FlagDebug enables debug logging
	FlagDebug Flag = "debug"
)
