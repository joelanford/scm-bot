package prefs

import (
	"github.com/joelanford/scm-bot/pkg/scm"
)

type Server struct {
	Cloner scm.Cloner
}

func NewServer(cloner scm.Cloner) *Server {
	return &Server{
		Cloner: cloner,
	}
}

func (s *Server) Run() error {
	return nil
}
