package envdb

var (
	// DevMode Development mode switch. If true
	// debug logging and serving assets from disk
	// is enabled.
	DevMode bool

	// TestMode
	TestMode bool

	// TimeFormat global time format string
	TimeFormat = "15:04:05"
)

const (
	// Name application name
	Name = "envdb"

	// Version application version number
	Version = "0.4.1"

	// DefaultServerPort the default tcp server port
	DefaultServerPort = 3636

	// DefaultWebServerPort the default web server port
	DefaultWebServerPort = 8080
)
