package docker

import (
	"fmt"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
	"github.com/denkhaus/magelib/shx"
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

// RemoveUntaggedImages removes untagged Docker images.
//
// No parameters.
// Returns an error if the operation fails.
func RemoveUntaggedImages() error {
	logging.Info("remove untagged docker images")
	p := pipe.Line(
		pipe.Exec("docker", "images"),
		pipe.Exec("grep", "none"),
		pipe.Exec("tr", "-s", " "),
		pipe.Exec("cut", "-d", " ", "-f", "3"),
		pipe.Exec("xargs", "-r", "docker", "rmi", "-f"),
	)

	return shx.RunPipeVerbose(p)
}

// ContainerNameByLabel gets the name of a Docker container by its label.
//
// Parameters: label (string) - the label of the Docker container.
// Returns: name (string) - the name of the Docker container, err (error) - an error if the operation fails.
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

// Build builds a Docker image.
//
// Parameters: moduleDir (string) - the directory of the module, tag (string) - the tag of the image.
// Returns: error - an error if the operation fails.
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

// BuildWithFile builds a Docker image using the specified Dockerfile and tags it with the given tag.
//
// Parameters:
// - moduleDir: the directory of the module.
// - dockerfilePath: the path to the Dockerfile.
// - tag: the tag to give to the Docker image.
//
// Returns:
// - error: an error if the operation fails.
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

// BuildWithArgs builds a Docker image with the specified build arguments and tags it with the given tag.
//
// Parameters: moduleDir (string) - the directory of the module, tag (string) - the tag of the image, args (magelib.ArgsMap) - the build arguments.
// Returns: error - an error if the operation fails.
func BuildWithArgs(moduleDir, tag string, args magelib.ArgsMap) error {
	err := magelib.InDirectory(moduleDir, func() error {
		logging.Infof("docker: build image %s with build args", tag)

		buildArgs := createBuildArgs(args)
		params := []string{"build", "--tag", tag}
		params = append(params, buildArgs...)
		params = append(params, ".")

		p := []interface{}{"params: "}
		for v := range params {
			p = append(p, v)
		}
		logging.Info(p...)

		output, err := Out(params...)
		logging.Println(output)

		if !strings.Contains(output, tag) {
			return errors.New("docker build doesn't finish correctly")
		}

		return err
	})

	return err
}

// ImageDigestLocal retrieves the digest of a Docker image with the given tag on the local machine.
//
// Parameters:
// - tag: the tag of the Docker image.
//
// Returns:
// - string: the digest of the Docker image.
// - error: an error if the operation fails.
func ImageDigestLocal(tag string) (string, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return "", errors.Wrap(err, "NewClientFromEnv")
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

// ImageDigestRemote retrieves the digest of a Docker image with the given tag remotely.
//
// Parameters:
// - tag (string): the tag of the Docker image.
//
// Returns:
// - string: the digest of the Docker image.
// - error: an error if the operation fails.
func ImageDigestRemote(tag string) (string, error) {
	mg.Deps(ensureCrane)
	return CraneDigestOut(tag)
}

// Push pushes a Docker image with the given tag.
//
// Parameters:
// - tag: the tag of the Docker image to push.
//
// Returns:
// - error: an error if the push operation fails.
func Push(tag string) error {
	logging.Infof("push image %s", tag)
	return sh.RunV("docker", "push", tag)
}

// PushCmd returns a magelib.Cmd that pushes a Docker image with the given tag.
//
// Parameters:
// - tag (string): the tag of the Docker image to push.
//
// Returns:
// - magelib.Cmd: a function that returns an error if the push operation fails.
func PushCmd(tag string) magelib.Cmd {
	return func() error {
		return Push(tag)
	}
}

// PushOnDemand pushes a Docker image with the given tag only if the local and remote digests do not match.
//
// Parameters:
// - tag (string): the tag of the Docker image.
//
// Returns:
// - error: an error if the push operation fails.
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

// PushOnDemandCmd returns a magelib.Cmd that pushes a Docker image with the given tag only if the local and remote digests do not match.
//
// Parameters:
// - tag (string): the tag of the Docker image.
//
// Returns:
// - magelib.Cmd: a function that returns an error if the push operation fails.
func PushOnDemandCmd(tag string) magelib.Cmd {
	return func() error {
		return PushOnDemand(tag)
	}
}

// IsImageAvailable checks if a Docker image is available on the local machine.
//
// Parameters:
// - imageName (string): the name of the Docker image to check.
//
// Returns:
// - bool: true if the image is available, false otherwise.
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

// RemoveLocalImage removes a local Docker image.
//
// Parameters:
// - imageName (string): the name of the Docker image to remove.
//
// Returns:
// - error: an error if the removal operation fails.
func RemoveLocalImage(imageName string) error {
	if IsImageAvailable(imageName) {
		logging.Infof("docker: remove local image %s", imageName)
		if err := sh.RunV("docker", "rmi", imageName); err != nil {
			return errors.Wrap(err, "docker [rmi]")
		}
	}

	return nil
}

// createBuildArgs generates a list of Docker build arguments from a map of key-value pairs.
//
// Parameters:
// - args (magelib.ArgsMap): a map of key-value pairs to be converted into build arguments.
//
// Returns:
// - []string: a list of build arguments in the format "--build-arg key=value".
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
