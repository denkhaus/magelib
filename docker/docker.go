package docker

import (
	"fmt"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	pipe "gopkg.in/pipe.v2"
)

var (
	Out            = sh.OutCmd("docker")
	CraneDigestOut = sh.OutCmd("crane", "digest")
)

// RemoveUntaggedImages as magelib.Cmd
func RemoveUntaggedImagesCmd() magelib.Cmd {
	return func() error {
		return RemoveUntaggedImages()
	}
}

func RemoveUntaggedImages() error {
	logging.Info("remove untagged docker images")
	p := pipe.Line(
		pipe.Exec("docker", "images"),
		pipe.Exec("grep", "none"),
		pipe.Exec("tr", "-s", " "),
		pipe.Exec("cut", "-d", " ", "-f", "3"),
		pipe.Exec("xargs", "-r", "docker", "rmi", "-f"),
	)

	output, err := pipe.CombinedOutput(p)
	if len(output) > 0 {
		logging.Info(string(output))
	}

	return err
}

func ContainerNameByLabel(label string) (string, error) {
	label = fmt.Sprintf("'label=%s'", label)
	name, err := Out(
		"ps",
		"-n", "1",
		"--filter", label,
		"--format", "'{{.Names}}'",
	)

	return name, err
}

// Build as magelib.Cmd
func BuildCmd(moduleDir, tag string) magelib.Cmd {
	return func() error {
		return Build(moduleDir, tag)
	}
}

func Build(moduleDir, tag string) error {
	err := magelib.InDirectory(moduleDir, func() error {
		logging.Infof("build image %s", tag)
		output, err := Out("build", "-t", tag, ".")
		fmt.Println(output)

		if !strings.Contains(output, tag) {
			return errors.New("docker build doesn't finish correctly")
		}

		return err
	})

	return err
}

// BuildWithFile as magelib.Cmd
func BuildWithFileCmd(moduleDir, dockerfilePath, tag string) magelib.Cmd {
	return func() error {
		return BuildWithFile(moduleDir, dockerfilePath, tag)
	}
}

func BuildWithFile(moduleDir, dockerfilePath, tag string) error {
	err := magelib.InDirectory(moduleDir, func() error {
		logging.Infof("build image %s", tag)
		return sh.RunV("docker", "build", "-t", tag, "-f", dockerfilePath, ".")
	})

	return err
}

// BuildWithArgs as magelib.Cmd
func BuildWithArgsCmd(moduleDir, tag string, args magelib.ArgsMap) magelib.Cmd {
	return func() error {
		return BuildWithArgs(moduleDir, tag, args)
	}
}

func BuildWithArgs(moduleDir, tag string, args magelib.ArgsMap) error {
	err := magelib.InDirectory(moduleDir, func() error {
		logging.Infof("docker: build image %s with build args", tag)

		buildArgs := createBuildArgs(args)
		params := []string{"build", "--tag", tag}
		params = append(params, buildArgs...)
		params = append(params, ".")

		logging.Info("params", params)
		output, err := Out(params...)
		logging.Println(output)

		if !strings.Contains(output, tag) {
			return errors.New("docker build doesn't finish correctly")
		}

		return err
	})

	return err
}

func ImageDigestLocal(tag string) (string, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return "", errors.Wrap(err, "NewClientWithOpts")
	}

	images, err := cli.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return "", errors.Wrap(err, "ListImages")
	}

	for _, image := range images {
		for _, t := range image.RepoTags {
			if t == tag {
				if len(image.RepoDigests) > 0 {
					digest := image.RepoDigests[0]
					if strings.Contains(digest, "@") {
						return strings.Split(digest, "@")[1], err
					}
				}

				// no digest localy
				return "", nil
			}
		}
	}

	return "", errors.Errorf("image %q not found", tag)
}

func ImageDigestRemote(tag string) (string, error) {
	mg.Deps(ensureCrane)
	return CraneDigestOut(tag)
}

func Push(tag string) error {
	logging.Infof("push image %s", tag)
	return sh.RunV("docker", "push", tag)
}

func PushCmd(tag string) magelib.Cmd {
	return func() error {
		return Push(tag)
	}
}

func PushOnDemand(tag string) error {
	digestLocal, err := ImageDigestLocal(tag)
	if err != nil {
		return errors.Wrap(err, "ImageDigestLocal")
	}

	digestRemote, err := ImageDigestRemote(tag)
	if err == nil && digestLocal == digestRemote {
		logging.Infof("remote image %s is in sync with local version", tag)
		return nil
	}

	return Push(tag)
}

func PushOnDemandCmd(tag string) magelib.Cmd {
	return func() error {
		return PushOnDemand(tag)
	}
}

func IsImageAvailable(imageName string) bool {
	if err := sh.Run("docker", "inspect", "--type", "image", imageName); err == nil {
		return true
	}

	return false
}

// RemoveLocalImage as magelib.Cmd
func RemoveLocalImageCmd(imageName string) magelib.Cmd {
	return func() error {
		return RemoveLocalImage(imageName)
	}
}

func RemoveLocalImage(imageName string) error {
	if IsImageAvailable(imageName) {
		logging.Infof("docker: remove local image %s", imageName)
		if err := sh.RunV("docker", "rmi", imageName); err != nil {
			return errors.Wrap(err, "DockerOut")
		}
	}

	return nil
}

func createBuildArgs(args magelib.ArgsMap) []string {
	out := []string{}

	if args == nil {
		return out
	}

	for key, value := range args {
		out = append(out, "--build-arg")
		out = append(out, fmt.Sprintf("%s=%s", key, value))
	}

	return out
}
