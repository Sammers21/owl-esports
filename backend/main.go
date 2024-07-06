package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
	"github.io/sammers21/owl-esports/backend/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	storage := dotabuff.NewStorage()
	server := http.NewServer(storage)
	go server.Start(8080)
	err := storage.LoadHeroes()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading heroes")
	}
	err = storage.LoadCounters()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading counters")
	}
	select {}
}
