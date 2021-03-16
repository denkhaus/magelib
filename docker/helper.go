package docker

import (
	"os/exec"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
)

func ensureCrane() error {
	if _, err := exec.LookPath("crane"); err != nil {
		logging.Info("install crane")
		return magelib.GoGet("github.com/google/go-containerregistry/cmd/crane")
	}

	return nil
}
