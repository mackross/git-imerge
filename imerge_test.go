package imerge

import (
	"fmt"
	//	"io/ioutil"
	//	"os"
	"testing"
	"time"

	git "github.com/libgit2/git2go"
)

/*
func TestFindCommit(t *testing.T) {
	repo, cleanup := createTestRepo(t)
	defer cleanup()
	// Create initial commit on master
	seedRepo(t, repo)

	// Add a commit on branch-1
	headCommit, err := getHeadCommit(repo)
	branch, err := repo.CreateBranch("branch-1", headCommit, false, signature, "Created a branch message")
	ok(t, err)
	masterCommitID := headCommit.Id()
	branch1CommitID, branch1TreeID := updateFileParentCommit(t, repo, headCommit, "README", "foo1", "Foo1 commit", branch.Reference.Name())

	// Test that branch name works
	commit, err := findCommit("branch-1", repo)
	ok(t, err)
	equals(t, branch1CommitID.String(), commit.Id().String())

	// Test that short commit sha is correct
	commit, err = findCommit("b1b3f", repo)
	ok(t, err)
	equals(t, branch1CommitID.String(), commit.Id().String())

	// Test HEAD can be found
	commit, err = findCommit("HEAD", repo)
	ok(t, err)
	equals(t, masterCommitID.String(), commit.Id().String())

	// Test branch-1 ref can be found
	commit, err = findCommit(branch.Reference.Name(), repo)
	ok(t, err)
	equals(t, branch1CommitID.String(), commit.Id().String())

	// Test non commit creates an error
	_, err = findCommit(branch1TreeID.String(), repo)
	assert(t, err != nil, "should only find commits")
}

func TestErrorOpeningNonExistantRepo(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	ok(t, err)

	err = os.Chdir(tmpdir)
	ok(t, err)

	_, err = RepoInCurrentDirectory()
	assert(t, err != nil, "no repo err")

	err = os.Remove(tmpdir)
	ok(t, err)
}

func TestGettingOursAndTheirs(t *testing.T) {
	repo, cleanup := createTestRepo(t)
	defer cleanup()
	// Create initial commit on master
	seedRepo(t, repo)

	// Add a commit on branch-1
	headCommit, err := getHeadCommit(repo)
	branch, err := repo.CreateBranch("branch-1", headCommit, false, signature, "Created a branch message")
	ok(t, err)
	branchCommit, _ := updateFileParentCommit(t, repo, headCommit, "README", "foo1", "Foo1 commit", branch.Reference.Name())

	// Test the correct commits are chosen
	ours, theirs, err := findOursAndTheirs("branch-1", repo)
	ok(t, err)

	equals(t, ours.Id().String(), headCommit.Id().String())
	equals(t, theirs.Id().String(), branchCommit.String())
}

func TestInvalidRevisionsReturnAnError(t *testing.T) {
	repo, cleanup := createTestRepo(t)
	defer cleanup()
	// Create initial commit on master
	seedRepo(t, repo)

	// Add a commit on branch-1
	headCommit, err := getHeadCommit(repo)
	branch, err := repo.CreateBranch("branch-1", headCommit, false, signature, "Created a branch message")
	ok(t, err)
	updateFileParentCommit(t, repo, headCommit, "README", "foo1", "Foo1 commit", branch.Reference.Name())

	err = Begin("branch-22", repo)
	assert(t, err != nil, "invalid branches should cause an error")
}

func TestInvalidMergesReturnErrors(t *testing.T) {
	repo, cleanup := createTestRepo(t)
	defer cleanup()
	// Create initial commit on master
	seedRepo(t, repo)

	// Add a commit on branch-1
	headCommit, err := getHeadCommit(repo)
	branch, err := repo.CreateBranch("branch-1", headCommit, false, signature, "Created a branch message")
	ok(t, err)
	updateFileParentCommit(t, repo, headCommit, "README", "foo1", "Foo1 commit", branch.Reference.Name())

	err = Begin("master", repo)
	assert(t, err != nil, "can't merge the same id")

	removeRefs(repo)

	err = Begin("branch-1", repo)
	assert(t, err != nil, "fast forwards should have an error")
}
*/

func TestRefsAreWritten(t *testing.T) {
	repo, cleanup := createTestRepo(t)
	defer cleanup()
	// Create initial commit on master
	seedRepo(t, repo)

	// Add a commit on branch-1
	headCommit, err := getHeadCommit(repo)
	branch, err := repo.CreateBranch("branch-1", headCommit, false, signature, "Created a branch message")
	ok(t, err)
	updateFileParentCommit(t, repo, headCommit, "README", "foo1", "Foo1 commit", branch.Reference.Name())
	updateFileParentCommit(t, repo, headCommit, "README", "foo2", "Foo2 commit", "HEAD")

	err = Begin("branch-1", repo)
	assert(t, len(imergeRefs(repo)) != 0, "imerge refs should exist")
}

func getHeadCommit(repo *git.Repository) (*git.Commit, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return repo.LookupCommit(head.Target())
}
func xestFastForwardReturnsErr(t *testing.T) {

	repo, cleanup := createTestRepo(t)
	defer cleanup()

	seedRepo(t, repo)

	head, err := repo.Head()
	ok(t, err)

	commit, err := repo.LookupCommit(head.Target())
	ok(t, err)

	updateFileParentCommit(t, repo, commit, "README", "foo1", "bar1", head.Name())

	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2015, 12, 12, 06, 20, 40, 0, time.UTC),
	}

	branch, err := repo.CreateBranch("b1", commit, false, sig, " a branch message")
	ok(t, err)

	repo.SetHead(branch.Reference.Name(), sig, "set head to b1")

	updateFileOnBranch(t, repo, "master", "README", "foo2", "bar2")

	emptyCommit(t, repo, "master", "test1")

	emptyCommit(t, repo, "b1", "test empty")
	updateFileOnBranch(t, repo, "b1", "README", "foo3", "bar3")
	emptyCommit(t, repo, "b1", "test2")
	emptyCommit(t, repo, "b1", "test3")
	updateFileOnBranch(t, repo, "master", "README", "foo4", "bar4")

	_, err = repo.CreateReference("refs/goimerge/test", branch.Target(), true, sig, "create new imerge ref")
	ok(t, err)

	itr, err := repo.NewReferenceIteratorGlob("refs/goimerge/*")
	ok(t, err)

	for {
		ref, err := itr.Next()
		if ref == nil {
			break
		}
		ok(t, err)
		fmt.Println(ref.Name())
	}

	err = repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})
	ok(t, err)

	idx, err := repo.Index()
	ok(t, err)

	b1Commit, err := branchCommit("b1", repo)
	ok(t, err)
	masterCommit, err := branchCommit("master", repo)
	ok(t, err)

	b1tree, err := b1Commit.Tree()
	ok(t, err)
	masterTree, err := masterCommit.Tree()
	ok(t, err)

	jointParent, err := repo.MergeBase(b1Commit.Id(), masterCommit.Id())
	ok(t, err)

	jointCommit, err := repo.LookupCommit(jointParent)
	ok(t, err)

	jointTree, err := jointCommit.Tree()
	ok(t, err)

	opts, err := git.DefaultMergeOptions()
	ok(t, err)

	idx, err = repo.MergeTrees(jointTree, b1tree, masterTree, &opts)
	ok(t, err)

	fmt.Println("Conflicts...", idx.HasConflicts())
	//	assert(t, ref != head, "test...", commitId, treeId)
}

func branchCommit(branch string, repo *git.Repository) (*git.Commit, error) {
	b, err := repo.LookupBranch(branch, git.BranchLocal)
	if err != nil {
		return nil, err
	}
	return repo.LookupCommit(b.Target())
}
func headCommit(t *testing.T, repo *git.Repository) *git.Commit {
	head, err := repo.Head()
	ok(t, err)

	commit, err := repo.LookupCommit(head.Target())
	ok(t, err)
	return commit
}
