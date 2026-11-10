// Package version reports the build version of KPass as a short git commit
// SHA. Go stamps VCS details into binaries built from a git checkout (both
// `go build` and the flatpak build), so no -ldflags plumbing is required.
package version

import "runtime/debug"

const (
	shortSHALen = 7
	// Dev is reported when no VCS information was stamped (e.g. `go run`).
	Dev = "dev"
)

// String returns the short commit SHA the binary was built from, suffixed
// with "-dirty" if the working tree had uncommitted changes, or Dev if the
// build carries no VCS information.
func String() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Dev
	}
	return fromSettings(info.Settings)
}

func fromSettings(settings []debug.BuildSetting) string {
	var revision string
	var modified bool
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return Dev
	}
	if len(revision) > shortSHALen {
		revision = revision[:shortSHALen]
	}
	if modified {
		revision += "-dirty"
	}
	return revision
}
