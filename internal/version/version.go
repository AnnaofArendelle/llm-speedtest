package version

import (
	"fmt"
	"runtime"
)

// Build information injected at link time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info is a serialisable description of the running binary.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the build information of the running binary.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a one-line human readable version string.
func String() string {
	return fmt.Sprintf("llm-speedtest %s (commit %s, built %s, %s %s/%s)", Version, Commit, Date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
