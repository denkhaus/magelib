package magelib

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4"

	"github.com/stretchr/testify/suite"
)

type magelibTest struct {
	suite.Suite
}

func (suite *magelibTest) SetupTest() {

}

func (suite *magelibTest) TearDownTest() {

}

func (suite *magelibTest) TestGitCommit() {
	dir, err := os.Getwd()
	if err != nil {
		suite.FailNow("PlainOpen")
	}

	r, err := git.PlainOpen(dir)
	if err != nil {
		suite.FailNow("PlainOpen")
	}

	head, err := r.Head()
	if err != nil {
		suite.FailNow("Head")
	}

	hash := head.Hash()

	commit, err := r.CommitObject(hash)
	if err != nil {
		suite.FailNow("CommitObject")
	}
	fmt.Println(commit.String())

	//suite.True(sa.Modified(), "StructAccessor modified")
	//suite.Equal(2, mediatorR.Audios.Count(), "MultiObject Count")
}

func TestCommon(t *testing.T) {
	testSuite := new(magelibTest)
	suite.Run(t, testSuite)
}
