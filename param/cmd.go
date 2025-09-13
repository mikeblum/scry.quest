// Package param provides CLI command and flag constants
package param

// Cmd represents CLI command types
type Cmd string

const (
	// CmdGenerate represents the generate command
	CmdGenerate Cmd = "generate"
	// CmdMigrate represents the migrate command
	CmdMigrate Cmd = "migrate"
)
