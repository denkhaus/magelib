package docker

import (
	"os/exec"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
)

func ensureCrane() error {
	if _, err := exec.LookPath("crane"); err != nil {
		logging.Info("install crane")
		return common.GoGet("github.com/google/go-containerregistry/cmd/...")
	}

	return nil
}
