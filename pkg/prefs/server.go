package prefs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joelanford/scm-bot/pkg/scm"
	"github.com/pkg/errors"
)

type Server struct {
	ListenAddress string
	Cloners       map[scm.RepoType]scm.Cloner
}

func NewServer(listenAddress string, cloners ...scm.Cloner) *Server {
	clonerMap := make(map[scm.RepoType]scm.Cloner)
	for _, cloner := range cloners {
		clonerMap[cloner.Type()] = cloner
	}
	return &Server{
		ListenAddress: listenAddress,
		Cloners:       clonerMap,
	}
}

type prefsResult struct {
	Status       string        `json:"status"`
	Error        string        `json:"error,omitempty"`
	CloneResults []cloneResult `json:"results,omitempty"`
}

type cloneResult struct {
	Success bool         `json:"success"`
	Error   string       `json:"error,omitempty"`
	URL     string       `json:"url"`
	Path    string       `json:"path"`
	Type    scm.RepoType `json:"type"`
}

func (s *Server) Run() error {
	http.HandleFunc("/preferences", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var (
			code   int
			result prefsResult
		)

		prefs, err := ReadFrom(r.Body)
		if err != nil {
			log.Println(errors.Wrapf(err, "server: could not read preferences"))

			code = http.StatusBadRequest
			result.Error = fmt.Sprintf("could not read preferences: %s", err)
			result.Status = "failed"
		} else {
			code = http.StatusOK
			cloneErrors := 0
			for _, repo := range prefs.GetRepositories() {
				cr := cloneResult{
					URL:  repo.URL,
					Path: repo.Path,
					Type: repo.Type,
				}
				cloner, ok := s.Cloners[repo.Type]
				if !ok {
					log.Println(errors.Errorf("server: no cloner for repo type \"%s\"", repo.Type))

					cr.Success = false
					cr.Error = fmt.Sprintf("no cloner for repo type")
					cloneErrors++
				} else if err := cloner.Clone(repo.Path, repo.URL); err != nil {
					log.Println(errors.Wrapf(err, "server: could not clone repo \"%s\" into path \"%s\"", repo.URL, repo.Path))

					cr.Success = false
					cr.Error = fmt.Sprintf("could not clone repo: %s", err)
					cloneErrors++
				} else {
					log.Printf("server: cloned repo \"%s\" into path \"%s\"", repo.URL, repo.Path)

					cr.Success = true
					cr.Error = ""
				}
				result.CloneResults = append(result.CloneResults, cr)
			}
			if cloneErrors == 0 {
				result.Status = "success"
			} else if cloneErrors == len(result.CloneResults) {
				result.Status = "failure"
			} else {
				result.Status = "partial success"
			}
		}
		data, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"success\":false, \"error\":\"%s\"}", err)))
		}
		w.WriteHeader(code)
		w.Write(data)
	})
	log.Printf("stated scm-bot preferences server at %s", s.ListenAddress)
	return http.ListenAndServe(s.ListenAddress, nil)
}
