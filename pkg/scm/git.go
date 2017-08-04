package scm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type GitCloner struct{}

func NewGitCloner() *GitCloner {
	return &GitCloner{}
}

func (c *GitCloner) Clone(into string, url string) error {
	auth, err := ssh.NewPublicKeysFromFile("git", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
	if err != nil {
		return errors.Wrap(err, "could not load SSH credentials")
	}

	repoDir := filepath.Join(into, strings.TrimSuffix(filepath.Base(url), ".git"))
	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		Auth:              auth,
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	return err
}

func (c *GitCloner) Type() RepoType {
	return GitRepoType
}
