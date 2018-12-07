package docker

import (
	"fmt"

	"bitbucket.org/denkhaus/metamage/lib/common"
	"github.com/magefile/mage/sh"
)

var (
	Out = sh.OutCmd("docker")
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
		return sh.RunV("docker", "build", "-t", tag, ".")
	})

	return err
}

func Push(tag string) error {
	return sh.RunV("docker", "push", tag)
}
