package scm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	git "gopkg.in/src-d/go-git.v4"
)

type GitCloner struct{}

func NewGitCloner() *GitCloner {
	return &GitCloner{}
}

func (c *GitCloner) Clone(into string, url string) CloneResult {
	cr := CloneResult{
		Success: true,
		Error:   nil,
		Path:    into,
		URL:     url,
		Type:    GitRepoType,
	}
	if url == "" {
		cr.Success = false
		cr.Error = errors.New("URL is undefined")
		return cr
	}

	repoPath := filepath.Join(into, strings.TrimSuffix(filepath.Base(url), ".git"))
	cr.Path = repoPath

	_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err == git.ErrRepositoryAlreadyExists {
		cr.Success = false
		cr.Error = ErrExists
	} else if err != nil {
		//
		// If there was any other clone error, cleanup the repo directory.
		//
		os.RemoveAll(repoPath)

		cr.Success = false
		cr.Error = err
	}

	return cr
}

func (c *GitCloner) Type() RepoType {
	return GitRepoType
}
