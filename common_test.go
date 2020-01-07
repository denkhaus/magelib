package magelib

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4"

	"github.com/stretchr/testify/suite"
)

type commonTest struct {
	suite.Suite
}

func (suite *commonTest) SetupTest() {

}

func (suite *commonTest) TearDownTest() {

}

func (suite *commonTest) TestStructAccessor() {
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
	testSuite := new(commonTest)
	suite.Run(t, testSuite)
}
