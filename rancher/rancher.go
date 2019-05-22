package rancher

import (
	"fmt"

	"github.com/denkhaus/magelib/common"
	"github.com/magefile/mage/sh"
)

var (
	RancherOut = sh.OutCmd("rancher")
	ComposeOut = sh.OutCmd("rancher-compose")
)

func ContainerNameByLabel(host, label string) string {
	label = fmt.Sprintf("label=%s", label)
	name, err := RancherOut(
		"--host", host,
		"docker", "ps",
		"-n", "1",
		"--filter", label,
		"--format", "{{.Names}}",
	)

	common.HandleError(err)
	return name
}

func RancherCompose(moduleDir, stack string) error {
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

func RancherComposeWith(env map[string]string, moduleDir, stack string) error {
	err := common.InDirectory(moduleDir, func() error {
		return common.RunVWith(env, "rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

func Rancher(moduleDir, stack string) error {
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher", "-s", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

func RancherWith(env map[string]string, moduleDir, stack string) error {
	err := common.InDirectory(moduleDir, func() error {
		return common.RunVWith(env, "rancher", "-s", stack, "up", "-d", "--force-upgrade")
	})

	return err
}
