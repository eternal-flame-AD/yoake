package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Version = func() string {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "unknown"
		}
		for _, setting := range info.Settings {
			fmt.Printf("setting: %s=%s", setting.Key, setting.Value)
		}
		return info.Main.Version
	}()
	BuildDate = "unknown"
)
