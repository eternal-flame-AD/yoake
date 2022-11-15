package version

import "runtime/debug"

var (
	Version = func() string {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "unknown"
		}
		return info.Main.Version
	}()
	BuildDate = "unknown"
)
