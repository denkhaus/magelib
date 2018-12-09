package docker

import (
	"fmt"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	"github.com/juju/errors"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Out            = sh.OutCmd("docker")
	CraneDigestOut = sh.OutCmd("crane", "digest")
)

func ContainerNameByLabel(label string) string {
	label = fmt.Sprintf("'label=%s'", label)
	name, err := Out(
		"ps",
		"-n", "1",
		"--filter", label,
		"--format", "'{{.Names}}'",
	)

	common.HandleError(err)
	return name
}

func Build(moduleDir, tag string) error {
	err := common.InDirectory(moduleDir, func() error {
		logging.Infof("build image %s", tag)
		return sh.RunV("docker", "build", "-t", tag, ".")
	})

	return err
}

func ImageDigestLocal(tag string) (string, error) {
	digest, err := Out("inspect",
		"--format", "{{index .RepoDigests 0}}",
		tag,
	)

	if strings.Contains(digest, "@") {
		return strings.Split(digest, "@")[1], err
	}

	return digest, err
}

func ImageDigestRemote(tag string) (string, error) {
	mg.Deps(ensureCrane)
	return CraneDigestOut(tag)
}

func Push(tag string) error {
	logging.Infof("push image %s", tag)
	return sh.RunV("docker", "push", tag)
}

func PushOnDemand(tag string) error {
	digestLocal, err := ImageDigestLocal(tag)
	if err != nil {
		return errors.Annotate(err, "ImageDigestLocal")
	}

	digestRemote, err := ImageDigestRemote(tag)
	if err == nil && digestLocal == digestRemote {
		logging.Infof("remote image %s is in sync with local version", tag)
		return nil
	}

	return Push(tag)
}
