package filestore

import "github.com/spf13/afero"

type FS interface {
	afero.Fs
}

func NewFS(basepath string) FS {
	bp := afero.NewBasePathFs(afero.NewOsFs(), basepath)
	return bp
}

func ChrootFS(fs FS, chroot string) FS {
	return afero.NewBasePathFs(fs, chroot)
}
