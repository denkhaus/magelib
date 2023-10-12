package docker

import (
	"os/exec"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
)

func ensureCrane() error {
	if _, err := exec.LookPath("crane"); err != nil {
		logging.Info("install crane")
		return magelib.GoInstall("github.com/google/go-containerregistry/cmd/crane@latest")
	}

	return nil
}
