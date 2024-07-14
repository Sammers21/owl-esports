package main

import (
	"flag"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

func main() {
	telegramTokenCli := flag.String("t", "", "Telegram bot token")
	mysqlCli := flag.String("m", "", "MySQL connection string")
	flag.Parse()
	telegramToken := *telegramTokenCli
	mysql := *mysqlCli
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	mysqlDb, err := dotabuff.NewMySQL(mysql)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to MySQL")
		return
	}
	engine := dotabuff.NewEngine(mysqlDb)
	var telegramBot *dotabuff.TelegramBot
	if telegramToken != "" {
		log.Info().Str("token", telegramToken).Msg("Starting telegram bot")
		telegramBot = dotabuff.NewTelegramBot(engine, telegramToken)
		go telegramBot.Start()
	} else {
		log.Info().Msg("Telegram bot token not provided")
	}
	server := dotabuff.NewServer(engine, telegramBot)
	go server.Start(8080)
	engineUpd := func() {
		log.Info().Msg("Updating heroes and counters data from dotabuff")
		err := engine.LoadHeroes()
		if err != nil {
			log.Fatal().Err(err).Msg("Error loading heroes")
		}
		err = engine.LoadCounters()
		if err != nil {
			log.Fatal().Err(err).Msg("Error loading counters")
		}
	}
	go func() {
		engineUpd()
		// update every 30 minutes
		for range time.Tick(30 * time.Minute) {
			engineUpd()
		}
	}()
	select {}
}
