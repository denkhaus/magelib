package git

import (
	"context"

	"github.com/denkhaus/magelib"
	"github.com/usvc/go-semver"
)

// NextVersion generates the next version based on the most recent tag and the maximum patch count.
//
// The ctx parameter is used to handle the context of the function call.
// The maxCount parameter specifies the maximum patch count before bumping the minor version.
// Returns the next version as a string and an error if any occurs.
func NextVersion(ctx context.Context, maxCount uint) (string, error) {
	tag, err := MostRecentTag()
	if err != nil {
		return "", magelib.Fatal(err, "MostRecentTag")
	}

	version := semver.Parse(tag)

	if version.VersionCore.Patch >= maxCount {
		version.VersionCore.Patch = 0
		version.BumpMinor()
		return version.String(), nil
	}

	if version.VersionCore.Minor >= maxCount {
		version.VersionCore.Minor = 0
		version.BumpMajor()
		return version.String(), nil
	}

	version.BumpPatch()
	return version.String(), nil
}
