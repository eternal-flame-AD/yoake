package funcmap

import "github.com/eternal-flame-AD/yoake/internal/version"

type V struct {
	Version   string
	BuildDate string
}

func Version() (*V, error) {
	return &V{
		Version:   version.Version,
		BuildDate: version.BuildDate,
	}, nil
}
