package git

import (
	"io"
	"path/filepath"
	"time"

	"github.com/denkhaus/logging"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GitRepository struct {
	path    string
	repoURL string
}

var Name = "Repo Maintainer"
var Email = "unknown@github"

func NewGitRepository(repoPath, repoURL string) (*GitRepository, error) {
	path, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, errors.Wrap(err, "Abs")
	}

	rep := GitRepository{
		path:    path,
		repoURL: repoURL,
	}

	return &rep, nil
}

func (p *GitRepository) Clone(w io.Writer) error {
	_, err := git.PlainClone(p.path, false, &git.CloneOptions{
		URL:      p.repoURL,
		Progress: w,
	})

	if err != nil {
		return errors.Wrap(err, "PlainClone")
	}

	return nil
}

func (p *GitRepository) CommitAll(message string) error {
	r, err := git.PlainOpen(p.path)
	if err != nil {
		return errors.Wrap(err, "PlainOpen")
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "Worktree")
	}

	status, err := w.Status()
	if err != nil {
		return errors.Wrap(err, "Status")
	}

	logging.Info(status)

	if err := w.AddGlob("./**/*"); err != nil {
		return errors.Wrap(err, "AddGlob")
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
		return errors.Wrap(err, "Commit")
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		return errors.Wrap(err, "CommitObject")
	}

	logging.Info(obj)
	return nil
}
