package meta

import (
	"time"
)

type Service struct {
	YouTube *YouTube
}

func NewMetaService() (s *Service) {
	s = &Service{}

	s.Initialize()

	return
}

func (s *Service) Initialize() {
	s.YouTube = NewYoutubeService()
}

// GetMetaForItem - Get meta information from
func (s *Service) GetMetaForItem(source string) (*Meta, error) {

	// TODO Decompile source and request meta from other providers

	return s.YouTube.GetMetaForItem(source)
}

type Meta struct {
	Identifier string
	Title      string
	Artist     string
	Album      string
	Source     string
	Duration   time.Duration
}
