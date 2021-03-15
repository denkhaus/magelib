package rancher

import (
	"fmt"
	"os/exec"

	pipe "gopkg.in/pipe.v2"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	//TODO:  make tis customizable
	VersionRancherCLI     = "v0.6.10"
	VersionRancherCompose = "v0.12.5"
)

var (
	RancherOut = sh.OutCmd("rancher")
	ComposeOut = sh.OutCmd("rancher-compose")
)

func RancherURL(version string) string {
	return fmt.Sprintf("https://releases.rancher.com/cli/%s/rancher-linux-amd64-%s.tar.gz", version, version)
}

func RancherComposeURL(version string) string {
	return fmt.Sprintf("https://releases.rancher.com/compose/%s/rancher-compose-linux-amd64-%s.tar.gz", version, version)
}

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

func EnsureRancher() error {
	if _, err := exec.LookPath("rancher"); err != nil {
		logging.Info("install rancher CLI")
		return InstallRancher(VersionRancherCLI)
	}

	return nil
}

func InstallRancher(version string) error {
	p := pipe.Script(
		pipe.Line(
			pipe.Exec("curl", RancherURL(version)),
			pipe.Exec("sudo", "tar", "--strip-components", "2", "-C", "/usr/bin/", "-xzf", "-"),
		),
		pipe.Exec("ls", "-la", "/usr/bin/rancher"),
	)

	output, err := pipe.CombinedOutput(p)
	if len(output) > 0 {
		logging.Info(string(output))
	}

	return err
}

func EnsureRancherCompose() error {
	if _, err := exec.LookPath("rancher-compose"); err != nil {
		logging.Info("install rancher-compose")
		return InstallRancherCompose(VersionRancherCompose)
	}

	return nil
}

func InstallRancherCompose(version string) error {
	p := pipe.Script(
		pipe.Line(
			pipe.Exec("curl", RancherComposeURL(version)),
			pipe.Exec("sudo", "tar", "--strip-components", "2", "-C", "/usr/bin/", "-xzf", "-"),
		),
		pipe.Exec("ls", "-la", "/usr/bin/rancher-compose"),
	)

	output, err := pipe.CombinedOutput(p)
	if len(output) > 0 {
		logging.Info(string(output))
	}

	return err
}

func RancherCompose(moduleDir, stack string) error {
	mg.Deps(EnsureRancherCompose)
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

func RancherComposeWith(env map[string]string, moduleDir, stack string) error {
	mg.Deps(EnsureRancherCompose)
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunWithV(env, "rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

func Rancher(moduleDir, stack string) error {
	mg.Deps(EnsureRancher)
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher", "up", "-s", stack, "-d", "--force-upgrade")
	})

	return err
}

func RancherWith(env map[string]string, moduleDir, stack string) error {
	mg.Deps(EnsureRancher)
	err := common.InDirectory(moduleDir, func() error {
		return sh.RunWithV(env, "rancher", "up", "-s", stack, "-d", "--force-upgrade")
	})

	return err
}
