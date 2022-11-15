package funcmap

import "github.com/eternal-flame-AD/yoake/internal/version"

type V struct {
	Version string
	Date    string
}

func Version() (*V, error) {
	return &V{
		Version: version.Version,
		Date:    version.Date,
	}, nil
}
