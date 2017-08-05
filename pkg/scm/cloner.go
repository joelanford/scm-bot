package scm

import (
	"errors"
	"path/filepath"
)

type RepoType string

const (
	GitRepoType RepoType = "git"
)

var (
	ErrNoCloner = errors.New("no cloner available for this repository type")
	ErrExists   = errors.New("repository already exists")
)

type Repository struct {
	Type RepoType `json:"type"`
	Path string   `json:"path"`
	URL  string   `json:"url"`
}

type Cloner interface {
	Type() RepoType
	Clone(path string, url string) CloneResult
}

type CloneResult struct {
	Success bool     `json:"success"`
	Error   error    `json:"error,omitempty"`
	URL     string   `json:"url"`
	Path    string   `json:"path"`
	Type    RepoType `json:"type"`
}

func CloneAll(baseDirectory string, cloners []Cloner, repos []Repository) <-chan CloneResult {
	resChan := make(chan CloneResult)

	clonerMap := make(map[RepoType]Cloner)
	for _, cloner := range cloners {
		clonerMap[cloner.Type()] = cloner
	}

	go func() {
		defer close(resChan)
		for _, repo := range repos {
			cloner, ok := clonerMap[repo.Type]
			if !ok {
				resChan <- CloneResult{
					Success: false,
					Error:   ErrNoCloner,
					URL:     repo.URL,
					Type:    repo.Type,
				}
				continue
			}

			fullPath := filepath.Join(baseDirectory, repo.Path)
			resChan <- cloner.Clone(fullPath, repo.URL)
		}
	}()
	return resChan
}
