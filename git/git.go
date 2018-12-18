package git

import (
	"bytes"
	"path/filepath"
	"time"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	"github.com/juju/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GitRepository struct {
	path    string
	repoURL string
}

var Name = "Repo Maintainer"
var Email = "unknown@github"

func NewGitRepository(repoPath, repoURL string) *GitRepository {
	path, err := filepath.Abs(repoPath)
	common.HandleError(err)

	rep := GitRepository{
		path:    path,
		repoURL: repoURL,
	}

	return &rep
}

func (p *GitRepository) Clone() (string, error) {
	output := bytes.NewBuffer(nil)
	_, err := git.PlainClone(p.path, false, &git.CloneOptions{
		URL:      p.repoURL,
		Progress: output,
	})

	return string(output.Bytes()), err
}

func (p *GitRepository) CommitAll(message string) error {
	r, err := git.PlainOpen(p.path)
	if err != nil {
		return errors.Annotate(err, "PlainOpen")
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Annotate(err, "Worktree")
	}

	status, err := w.Status()
	if err != nil {
		return errors.Annotate(err, "Status")
	}

	logging.Info(status)

	if err := w.AddGlob("./**/*"); err != nil {
		return errors.Annotate(err, "AddGlob")
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  Name,
			Email: Email,
			When:  time.Now(),
		},
	})

	if err != nil {
		return errors.Annotate(err, "Commit")
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		return errors.Annotate(err, "CommitObject")
	}

	logging.Info(obj)
	return nil
}

func GitStatus(path string) (*StatusInfo, error) {
	gitOut, err := GitStatusOutput(path)
	if err != nil {
		return nil, errors.Annotate(err, "GitStatusOutput")
	}

	info := NewStatusInfo(path)
	if err := info.parseStatusOutput(gitOut); err != nil {
		return nil, errors.Annotate(err, "ParseStatusOutput")
	}

	return info, nil
}
