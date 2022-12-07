//go:build !tinygo

package version

import (
	"runtime/debug"
	"time"
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		Version = "unknown"
		return
	}

	var vcsRevision string
	var vcsTime time.Time
	var vcsModified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = setting.Value
		case "vcs.time":
			vcsTime, _ = time.Parse(time.RFC3339, setting.Value)
		case "vcs.modified":
			vcsModified = setting.Value != "false"
		}
	}
	if tagVersion != "" {
		vcsRevision = tagVersion
	}

	if vcsModified {
		Date = buildDate
		Version = vcsRevision + "+devel"
	} else {
		Date = buildDate
		if !vcsTime.IsZero() {
			Date = vcsTime.Format("2006-01-02T15:04Z07:00")
		}
		Version = vcsRevision
	}
}
