package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

type Storage struct {
	Heroes []*dotabuff.Hero
	Counters map[string][]*dotabuff.Counter
}

func newStorage() *Storage {
	return &Storage{
		Heroes: make([]*dotabuff.Hero, 0),
		Counters: make(map[string][]*dotabuff.Counter),
	}
}

func (s *Storage) LoadHeroes() error {
	log.Info().Msg("Loading heroes...")
	heroes, err := dotabuff.Heroes()
	if err != nil {
		return err
	}
	log.Info().
		Int("count", len(heroes)).
		Msg("Heroes has been loaded")
	s.Heroes = heroes
	return nil
}

func (s *Storage) LoadCounters() error {
	for i, hero := range s.Heroes {
		counters, err := hero.Counters()
		if err != nil {
			log.Printf("Error fetching counters for %s: %v", hero.Name, err)
			continue
		}
		log.Info().Msgf("%d/%d: %s has %d counters", i+1, len(s.Heroes), hero.Name, len(counters))
		s.Counters[hero.Name] = counters
	}
	return nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	storage := newStorage()
	err := storage.LoadHeroes()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading heroes")
	}
	err = storage.LoadCounters()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading counters")
	}
}
