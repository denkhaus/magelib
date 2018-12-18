package git

// from https://github.com/robertgzr/porcelain.git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/juju/errors"
)

const notRepoStatus string = "exit status 128"

var ErrNotAGitRepo error = errors.New("not a git repo")

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

	branch   string
	commit   string
	remote   string
	upstream string
	ahead    int
	behind   int

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

func (pi *StatusInfo) IsModified() bool {
	return pi.Unstaged.hasChanged()
}

func (pi *StatusInfo) IsDirty() bool {
	return pi.Staged.hasChanged()
}

func (pi *StatusInfo) Debug() string {
	return fmt.Sprintf("%#+v", pi)
}

func (pi *StatusInfo) ParseStatusOutput(r io.Reader) error {
	var s = bufio.NewScanner(r)

	for s.Scan() {
		if len(s.Text()) < 1 {
			continue
		}

		pi.ParseLine(s.Text())
	}

	return nil
}

func (pi *StatusInfo) ParseLine(line string) error {
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
//
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
		return nil, errors.Annotate(err, "IsInsideWorkTree")
	} else if !ok {
		return nil, ErrNotAGitRepo
	}

	var buf = new(bytes.Buffer)
	cmd := exec.Command("git", "status", "--porcelain=v2", "--branch")
	cmd.Stdout = buf
	cmd.Dir = cwd

	if err := cmd.Run(); err != nil {
		return nil, errors.Annotate(err, "run git status")
	}

	return buf, nil
}

func PathToGitDir(cwd string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--absolute-git-dir")
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Annotate(err, "run git rev-parse")
	}

	return strings.TrimSpace(string(out)), nil
}

func IsInsideWorkTree(cwd string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
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

		return false, err
	}

	return strconv.ParseBool(strings.TrimSpace(string(out)))
}

func consumeNext(s *bufio.Scanner) string {
	if s.Scan() {
		return s.Text()
	}
	return ""
}
