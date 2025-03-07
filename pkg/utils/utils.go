package utils

// Version information set during build by the Makefile
var (
	// Version is the semantic version of the application
	Version = "0.0.2"

	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"

	// BuildDate is the date when the application was built
	BuildDate = "unknown"
)

// GetVersionInfo returns a formatted string with the version information
func GetVersionInfo() string {
	return "Unregex " + Version + " (" + GitCommit + ") built on " + BuildDate
}

// Description returns a short description of the application
func Description() string {
	return "Work with temporal"
}
