package dmi

import (
	"strings"

	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/spf13/afero"
)

type DMI struct {
	log logger.Logger
	fs  afero.Fs
}

func New(log logger.Logger) *DMI {
	return &DMI{
		log: log,
		fs:  afero.NewOsFs(),
	}
}

func (d *DMI) read(path string) (string, error) {
	_, err := d.fs.Stat(path)
	if err != nil {
		return "", err
	}

	content, err := afero.ReadFile(d.fs, path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
