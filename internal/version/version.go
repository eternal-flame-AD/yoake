package version

import (
	"runtime/debug"
	"time"
)

var (
	tagVersion = ""
	buildDate  = "unknown"

	Date    = "unknown"
	Version = func() string {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "unknown"
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
			return vcsRevision + "+devel"
		} else {
			Date = buildDate
			if !vcsTime.IsZero() {
				Date = vcsTime.Format("2006-01-02T15:04Z07:00")
			}
			return vcsRevision
		}

	}()
)
