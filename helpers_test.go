package imerge

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	git "github.com/libgit2/git2go"
)

var signature = &git.Signature{
	Name:  "Rand Om Hacker",
	Email: "random@hacker.com",
	When:  time.Date(2015, 12, 12, 06, 20, 40, 0, time.UTC),
}

func emptyCommit(t *testing.T, repo *git.Repository, branchName string, msg string) {
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2015, 12, 12, 06, 20, 40, 0, time.UTC),
	}

	currentBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	ok(t, err)

	currentTip, err := repo.LookupCommit(currentBranch.Target())
	ok(t, err)

	tree, err := currentTip.Tree()
	ok(t, err)

	_, err = repo.CreateCommit(currentBranch.Reference.Name(), sig, sig, "Empty: "+msg, tree, currentTip)
	ok(t, err)
}

func updateFileOnBranch(t *testing.T, repo *git.Repository, branchName string, pathFromWorkdir string, content string, msg string) (*git.Oid, *git.Oid) {
	currentBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	ok(t, err)

	commit, err := repo.LookupCommit(currentBranch.Target())
	ok(t, err)

	return updateFileParentCommit(t, repo, commit, pathFromWorkdir, content, msg, currentBranch.Reference.Name())
}

func updateFileParentCommit(t *testing.T, repo *git.Repository, commit *git.Commit, pathFromWorkdir string, content string, msg string, ref string) (*git.Oid, *git.Oid) {
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2015, 12, 12, 06, 20, 40, 0, time.UTC),
	}

	idx, err := repo.Index()
	ok(t, err)

	pathDir := path.Join(path.Dir(path.Dir(repo.Path())), pathFromWorkdir)
	err = ioutil.WriteFile(pathDir, []byte(content), 0644)
	ok(t, err)

	err = idx.AddByPath(pathFromWorkdir)
	ok(t, err)

	err = idx.Write()
	ok(t, err)

	treeId, err := idx.WriteTree()
	ok(t, err)

	tree, err := repo.LookupTree(treeId)
	ok(t, err)

	commitId, err := repo.CreateCommit(ref, sig, sig, msg, tree, commit)
	ok(t, err)

	return commitId, treeId
}

func createTestRepo(t *testing.T) (*git.Repository, func()) {
	path, err := os.Getwd()
	ok(t, err)
	path = path + "/test"
	os.RemoveAll(path)

	repo, err := git.InitRepository(path, false)
	ok(t, err)

	return repo, func() {
		repo.Free()
	}
}

func seedRepo(t *testing.T, repo *git.Repository) {
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2015, 12, 12, 06, 20, 40, 0, time.UTC),
	}
	idx, err := repo.Index()
	ok(t, err)

	treeId, err := idx.WriteTree()
	ok(t, err)

	tree, err := repo.LookupTree(treeId)
	ok(t, err)

	_, err = repo.CreateCommit("HEAD", sig, sig, "first commit", tree)
	ok(t, err)
}

func repoWorkDir(repo *git.Repository) string {
	return path.Dir(path.Dir(repo.Path()))
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
