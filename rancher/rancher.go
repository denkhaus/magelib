package rancher

import (
	"fmt"
	"os/exec"

	pipe "gopkg.in/pipe.v2"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	//TODO:  make tis customizable
	VersionRancherCLI     = "v0.6.10"
	VersionRancherCompose = "v0.12.5"
)

var (
	Out        = sh.OutCmd("rancher")
	ComposeOut = sh.OutCmd("rancher-compose")
)

func RancherURL(version string) string {
	return fmt.Sprintf("https://releases.rancher.com/cli/%s/rancher-linux-amd64-%s.tar.gz", version, version)
}

func ComposeURL(version string) string {
	return fmt.Sprintf("https://releases.rancher.com/compose/%s/rancher-compose-linux-amd64-%s.tar.gz", version, version)
}

// EnsureRancher as magelib.Cmd
func EnsureRancherCmd() magelib.Cmd {
	return func() error {
		return EnsureRancher()
	}
}

func EnsureRancher() error {
	if _, err := exec.LookPath("rancher"); err != nil {
		logging.Info("install rancher CLI")
		return InstallRancher(VersionRancherCLI)
	}

	return nil
}

func ContainerNameByLabel(host, label string) (string, error) {
	label = fmt.Sprintf("label=%s", label)
	name, err := Out(
		"--host", host,
		"docker", "ps",
		"-n", "1",
		"--filter", label,
		"--format", "{{.Names}}",
	)

	return name, err
}

// InstallRancher as magelib.Cmd
func InstallRancherCmd(version string) magelib.Cmd {
	return func() error {
		return InstallRancher(version)
	}
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

// EnsureRancherCompose as magelib.Cmd
func EnsureComposeCmd() magelib.Cmd {
	return func() error {
		return EnsureCompose()
	}
}

func EnsureCompose() error {
	if _, err := exec.LookPath("rancher-compose"); err != nil {
		logging.Info("install rancher-compose")
		return InstallCompose(VersionRancherCompose)
	}

	return nil
}

// InstallCompose as magelib.Cmd
func InstallComposeCmd(version string) magelib.Cmd {
	return func() error {
		return InstallCompose(version)
	}
}

func InstallCompose(version string) error {
	p := pipe.Script(
		pipe.Line(
			pipe.Exec("curl", ComposeURL(version)),
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

// Compose as magelib.Cmd
func ComposeCmd(moduleDir, stack string) magelib.Cmd {
	return func() error {
		return Compose(moduleDir, stack)
	}
}

func Compose(moduleDir, stack string) error {
	mg.Deps(EnsureCompose)
	err := magelib.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

// ComposeWith as magelib.Cmd
func ComposeWithCmd(env magelib.ArgsMap, moduleDir, stack string) magelib.Cmd {
	return func() error {
		return ComposeWith(env, moduleDir, stack)
	}
}

func ComposeWith(env magelib.ArgsMap, moduleDir, stack string) error {
	mg.Deps(EnsureCompose)
	err := magelib.InDirectory(moduleDir, func() error {
		return sh.RunWithV(env, "rancher-compose", "-p", stack, "up", "-d", "--force-upgrade")
	})

	return err
}

// Rancher as magelib.Cmd
func RancherCmd(moduleDir, stack string) magelib.Cmd {
	return func() error {
		return Rancher(moduleDir, stack)
	}
}

func Rancher(moduleDir, stack string) error {
	mg.Deps(EnsureRancher)
	err := magelib.InDirectory(moduleDir, func() error {
		return sh.RunV("rancher", "up", "-s", stack, "-d", "--force-upgrade")
	})

	return err
}

// RancherWith as magelib.Cmd
func RancherWithCmd(env magelib.ArgsMap, moduleDir, stack string) magelib.Cmd {
	return func() error {
		return RancherWith(env, moduleDir, stack)
	}
}

func RancherWith(env magelib.ArgsMap, moduleDir, stack string) error {
	mg.Deps(EnsureRancher)
	err := magelib.InDirectory(moduleDir, func() error {
		return sh.RunWithV(env, "rancher", "up", "-s", stack, "-d", "--force-upgrade")
	})

	return err
}
