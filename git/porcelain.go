package git

// based on https://github.com/robertgzr/porcelain.git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const notRepoStatus string = "exit status 128"

var ErrNotAGitRepo = errors.New("not a git repo")

type GitArea struct {
	modified int
	added    int
	deleted  int
	renamed  int
	copied   int
}

func (a *GitArea) hasChanged() bool {
	var changed bool
	if a.added != 0 {
		changed = true
	}
	if a.deleted != 0 {
		changed = true
	}
	if a.modified != 0 {
		changed = true
	}
	if a.copied != 0 {
		changed = true
	}
	if a.renamed != 0 {
		changed = true
	}
	return changed
}

type StatusInfo struct {
	workingDir string
	branch     string
	commit     string
	remote     string
	upstream   string

	ahead     int
	behind    int
	untracked int
	unmerged  int

	Unstaged GitArea
	Staged   GitArea
}

func NewStatusInfo(path string) *StatusInfo {
	pi := StatusInfo{
		workingDir: path,
	}

	return &pi
}

// IsDirty returns true if unstaged files have changed
func (pi *StatusInfo) IsModified() bool {
	return pi.Unstaged.hasChanged()
}

// IsDirty returns true if staged files have changed
func (pi *StatusInfo) IsDirty() bool {
	return pi.Staged.hasChanged()
}

// IsSynced returns true if repo is in sync with remote
func (pi *StatusInfo) IsSynced() bool {
	return pi.ahead == 0 && pi.behind == 0
}

// Debug retrieves StatusInfo as string
func (pi *StatusInfo) Debug() string {
	return fmt.Sprintf("%#+v", pi)
}

func (pi *StatusInfo) parseStatusOutput(r io.Reader) error {
	var s = bufio.NewScanner(r)

	for s.Scan() {
		if len(s.Text()) < 1 {
			continue
		}

		pi.parseLine(s.Text())
	}

	return nil
}

func (pi *StatusInfo) parseLine(line string) error {
	s := bufio.NewScanner(strings.NewReader(line))
	// switch to a word based scanner
	s.Split(bufio.ScanWords)

	for s.Scan() {
		switch s.Text() {
		case "#":
			pi.parseBranchInfo(s)
		case "1":
			pi.parseTrackedFile(s)
		case "2":
			pi.parseRenamedFile(s)
		case "u":
			pi.unmerged++
		case "?":
			pi.untracked++
		}
	}
	return nil
}

func (pi *StatusInfo) parseBranchInfo(s *bufio.Scanner) (err error) {
	// uses the word based scanner from ParseLine
	for s.Scan() {
		switch s.Text() {
		case "branch.oid":
			pi.commit = consumeNext(s)
		case "branch.head":
			pi.branch = consumeNext(s)
		case "branch.upstream":
			pi.upstream = consumeNext(s)
		case "branch.ab":
			err = pi.parseAheadBehind(s)
		}
	}
	return err
}

func (pi *StatusInfo) parseAheadBehind(s *bufio.Scanner) error {
	// uses the word based scanner from ParseLine
	for s.Scan() {
		i, err := strconv.Atoi(s.Text()[1:])
		if err != nil {
			return err
		}

		switch s.Text()[:1] {
		case "+":
			pi.ahead = i
		case "-":
			pi.behind = i
		}
	}
	return nil
}

// parseTrackedFile parses the porcelain v2 output for tracked entries
// doc: https://git-scm.com/docs/git-status#_changed_tracked_entries
func (pi *StatusInfo) parseTrackedFile(s *bufio.Scanner) error {
	// uses the word based scanner from ParseLine
	var index int
	for s.Scan() {
		switch index {
		case 0: // xy
			pi.parseXY(s.Text())
		default:
			continue
			// case 1: // sub
			// 	if s.Text() != "N..." {
			// 		log.Println("is submodule!!!")
			// 	}
			// case 2: // mH - octal file mode in HEAD
			// 	log.Println(index, s.Text())
			// case 3: // mI - octal file mode in index
			// 	log.Println(index, s.Text())
			// case 4: // mW - octal file mode in worktree
			// 	log.Println(index, s.Text())
			// case 5: // hH - object name in HEAD
			// 	log.Println(index, s.Text())
			// case 6: // hI - object name in index
			// 	log.Println(index, s.Text())
			// case 7: // path
			// 	log.Println(index, s.Text())
		}
		index++
	}
	return nil
}

func (pi *StatusInfo) parseXY(xy string) error {
	switch xy[:1] { // parse staged
	case "M":
		pi.Staged.modified++
	case "A":
		pi.Staged.added++
	case "D":
		pi.Staged.deleted++
	case "R":
		pi.Staged.renamed++
	case "C":
		pi.Staged.copied++
	}

	switch xy[1:] { // parse unstaged
	case "M":
		pi.Unstaged.modified++
	case "A":
		pi.Unstaged.added++
	case "D":
		pi.Unstaged.deleted++
	case "R":
		pi.Unstaged.renamed++
	case "C":
		pi.Unstaged.copied++
	}
	return nil
}

func (pi *StatusInfo) parseRenamedFile(s *bufio.Scanner) error {
	return pi.parseTrackedFile(s)
}

func GitStatusOutput(cwd string) (io.Reader, error) {
	if ok, err := IsInsideWorkTree(cwd); err != nil {
		if err == ErrNotAGitRepo {
			return nil, ErrNotAGitRepo
		}
		return nil, errors.Wrap(err, "IsInsideWorkTree")
	} else if !ok {
		return nil, ErrNotAGitRepo
	}

	var stderr = new(bytes.Buffer)
	var stdout = new(bytes.Buffer)

	cmd := exec.Command("git", "status", "--porcelain=v2", "--branch")

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = cwd

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err,
			"git status --porcelain=v2 --branch err: [%s]",
			stderr.String(),
		)
	}

	return stdout, nil
}

func PathToGitDir(cwd string) (string, error) {
	var stderr = new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--absolute-git-dir")
	cmd.Stderr = stderr
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err,
			"git rev-parse --absolute-git-dir err: [%s]", stderr.String())
	}

	return strings.TrimSpace(string(out)), nil
}

func IsInsideWorkTree(cwd string) (bool, error) {
	var stderr = new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = stderr
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 128 {
					return false, ErrNotAGitRepo
				}
			}
		}

		if cmd.ProcessState.String() == notRepoStatus {
			return false, ErrNotAGitRepo
		}

		return false, errors.Wrapf(err,
			"git rev-parse --is-inside-work-tree err: [%s]",
			stderr.String(),
		)
	}

	return strconv.ParseBool(strings.TrimSpace(string(out)))
}

func consumeNext(s *bufio.Scanner) string {
	if s.Scan() {
		return s.Text()
	}
	return ""
}
