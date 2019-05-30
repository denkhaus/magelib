package docker

import (
	"fmt"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/juju/errors"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	pipe "gopkg.in/pipe.v2"
)

var (
	Out            = sh.OutCmd("docker")
	CraneDigestOut = sh.OutCmd("crane", "digest")
)

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

func BuildFunc(moduleDir, tag string) func() error {
	return func() error {
		return Build(moduleDir, tag)
	}
}

func Build(moduleDir, tag string) error {
	err := common.InDirectory(moduleDir, func() error {
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

func BuildWithFileFunc(moduleDir, dockerfilePath, tag string) func() error {
	return func() error {
		return BuildWithFile(moduleDir, dockerfilePath, tag)
	}
}

func BuildWithFile(moduleDir, dockerfilePath, tag string) error {
	err := common.InDirectory(moduleDir, func() error {
		logging.Infof("build image %s", tag)
		return sh.RunV("docker", "build", "-t", tag, "-f", dockerfilePath, ".")
	})

	return err
}

func ImageDigestLocal(tag string) (string, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return "", errors.Annotate(err, "NewClientWithOpts")
	}

	images, err := cli.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return "", errors.Annotate(err, "ListImages")
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
