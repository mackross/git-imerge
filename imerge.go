package imerge

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"time"

	git "github.com/libgit2/git2go"
)

func RepoInCurrentDirectory() (*git.Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return git.OpenRepository(dir)
}

func Begin(spec string, repo *git.Repository) error {
	ours, theirs, err := findOursAndTheirs(spec, repo)
	if err != nil {
		return err
	}
	err = checkMergeForErrors(ours, theirs, repo)
	if err != nil {
		return err
	}
	if anyImergeRefExists(repo) {
		return errors.New("Cannot begin a new merge while there is one in progress.")
	}
	saveTargetRefs(ours.Id(), theirs.Id(), repo)
	return nil
}

func imergeRefs(repo *git.Repository) map[string]*git.Reference {
	itr, err := repo.NewReferenceIteratorGlob("refs/goimerge/*")
	checkErr(err)

	refs := make(map[string]*git.Reference, 0)
	for {
		ref, err := itr.Next()
		if gerr, ok := err.(*git.GitError); ok && gerr.Code == git.ErrIterOver {
			break
		}
		checkErr(err)
		refs[ref.Name()] = ref
	}
	return refs
}

func Abort(repo *git.Repository) (err error) {
	refs := imergeRefs(repo)
	for _, v := range refs {
		delErr := v.Delete()
		if delErr != nil {
			err = delErr
		}
	}
	if anyImergeRefExists(repo) {
		return err
	}
	return nil
}

func insertMergeRefs(repo *git.Repository) error {
	return nil
}
func anyImergeRefExists(repo *git.Repository) bool {
	return len(imergeRefs(repo)) > 0
}

func removeRefs(repo *git.Repository) {

}

func signatureFromConfig(repo *git.Repository) *git.Signature {
	config, err := repo.Config()
	checkErr(err)

	name, err := config.LookupString("user.name")
	if err != nil {
		u, err := user.Current()
		checkErr(err)
		name = u.Name
	}

	email, err := config.LookupString("user.email")
	checkErr(err)

	return &git.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}
func saveTargetRefs(ours *git.Oid, theirs *git.Oid, repo *git.Repository) {
	sig := signatureFromConfig(repo)
	fmt.Println(sig.Email)

	_, err := repo.CreateReference("refs/goimerge/ours", ours, false, sig, "Create refs/goimerge/ours.")
	checkErr(err)
	_, err = repo.CreateReference("refs/goimerge/theirs", theirs, false, sig, "Create refs/goimerge/theirs.")
	checkErr(err)

	commonAncestor, err := repo.MergeBase(ours, theirs)
	checkErr(err)

	_, err = repo.CreateReference("refs/goimerge/ca", commonAncestor, false, sig, "Create refs/goimerge/theirs.")

	commit, err := repo.LookupCommit(ours)
	checkErr(err)

	commits := []*git.Commit{commit}

	walk, err := repo.Walk()
	walk.PushRange(commonAncestor.String() + ".." + ours.String())

}
func checkMergeForErrors(ours *git.Commit, theirs *git.Commit, repo *git.Repository) error {
	// check ours is equal to the HEAD commit
	head, err := repo.Head()
	checkErr(err)

	if !ours.Id().Equal(head.Target()) {
		panic("Ours commit arg is not equal to head commit")
	}

	// check we're not merging the same commit
	if ours.Id().Equal(theirs.Id()) {
		return errors.New(fmt.Sprintf("Cannot merge the same commit (%v).", ours.Id().String()))
	}

	// check that a normal merge is the only realistic option
	theirsA, err := repo.LookupAnnotatedCommit(theirs.Id())
	checkErr(err)

	analysis, prefs, err := repo.MergeAnalysis([]*git.AnnotatedCommit{theirsA})
	if analysis == git.MergeAnalysisUpToDate {
		return errors.New("Cannot merge a parent commit.")
	}
	if analysis == git.MergeAnalysisUnborn {
		return errors.New("Cannot merge unborn HEAD.")
	}
	if prefs == git.MergePreferenceFastForwardOnly {
		return errors.New("Cannot perform a merge between the commits as user preferences only allow fast forwards.")
	}
	if analysis&git.MergeAnalysisNormal == 0 {
		return errors.New("Cannot perform a merge between the commits.")
	}
	if analysis&git.MergeAnalysisFastForward == git.MergeAnalysisFastForward {
		return errors.New("Cannot perform a merge between the commits as fast forward merge is available.")
	}
	return nil
}

func findCommit(spec string, repo *git.Repository) (*git.Commit, error) {
	obj, err := repo.RevparseSingle(spec)
	if err != nil {
		return nil, err
	}
	if obj.Type() != git.ObjectCommit {
		return nil, errors.New("git revision specified an object that is not a commit")
	}
	return obj.(*git.Commit), nil
}

func findOursAndTheirs(spec string, repo *git.Repository) (*git.Commit, *git.Commit, error) {
	theirs, err := findCommit(spec, repo)
	if err != nil {
		return nil, nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, nil, err
	}

	ours, err := repo.LookupCommit(ref.Target())
	if err != nil {
		return nil, nil, err
	}

	return ours, theirs, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
