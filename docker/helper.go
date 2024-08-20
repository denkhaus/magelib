package docker

import (
	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
	"github.com/denkhaus/magelib/shx"
)

// ensureCrane Installs the crane tool if it's not already installed.
//
// No parameters.
// Returns an error if the installation fails.
func ensureCrane() error {
	if _, err := shx.IsAppInstalled("crane"); err != nil {
		logging.Info("install crane")
		return magelib.GoInstall("github.com/google/go-containerregistry/cmd/crane@latest")
	}

	return nil
}
