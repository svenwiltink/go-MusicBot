package util

var (
	// BuildDate - date and time at which the binary is build - imported through ldflags
	BuildDate string

	// BuildHost - Host on which the binary is built - imported through ldflags
	BuildHost string

	// GitCommit - git commit hash - imported through ldflags
	GitCommit string

	// GoVersion - Version of go this binary is built with - imported through ldflags
	GoVersion string

	// VersionTag - version tag
	VersionTag string
)
