package scm

type RepoType string

const (
	GitRepoType RepoType = "git"
)

type Repository struct {
	Type RepoType `json:"type"`
	Path string   `json:"path"`
	URL  string   `json:"url"`
}

type Cloner interface {
	Type() RepoType
	Clone(path string, url string) error
}
