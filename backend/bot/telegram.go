package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

type TelegramBot struct {
	Engine *dotabuff.Engine
	Token  string
}

func NewTelegramBot(engine *dotabuff.Engine, token string) *TelegramBot {
	return &TelegramBot{
		Engine: engine,
		Token:  token,
	}
}

type TGLogger struct{}

func (l *TGLogger) Printf(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}

func (l *TGLogger) Println(v ...interface{}) {
	log.Info().Msg(fmt.Sprint(v...))
}

func (b *TelegramBot) Start() error {
	log.Info().Msg("Starting telegram bot...")
	bot, err := tgbotapi.NewBotAPI(b.Token)
	tgbotapi.SetLogger(&TGLogger{})
	if err != nil {
		log.Error().Err(err).Msgf("Error creating telegram bot with token %s", b.Token)
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			text := update.Message.Text
			split := strings.Split(text, " ")
			log.Info().Str("username", update.Message.From.UserName).Str("text", text).Msg("Received message")
			if text == "/start" || text == "/help" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a bot that can help you with Dota 2 hero counters. To get started, type 10 hero names, 5 for each team, and I'll tell you the winrate for each team. For example, type 'muerta es beastmaster tiny sd gyro snapfire underlord hoodwink cm'.")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			} else if len(split) == 10 {
				rw, dw, err := b.Engine.PickWinRateFromLines(split)
				if err != nil {
					log.Error().Err(err).Msg("Error fetching pick winrate")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error fetching pick winrate: %v", err))
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Radiant winrate: %.2f%%\nDire winrate: %.2f%%", rw, dw))
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			} else if strings.HasPrefix(text, "https://www.dotabuff.com/matches/") {
				match, err := dotabuff.ExtractHerosFromDBLink(text)
				if err != nil {
					log.Error().Err(err).Msg("Error extracting heroes from match")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error extracting heroes from match: %v", err))
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					continue
				}
				rw, dw, err := b.Engine.PickWinRateFromDBMatch(match)
				if err != nil {
					log.Error().Err(err).Msg("Error fetching pick winrate")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error fetching pick winrate: %v", err))
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Radiant winrate: %.2f%%\nDire winrate: %.2f%%", rw, dw))
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I'm sorry, I didn't understand that. Please type /start or /help for instructions.")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
	panic("unreachable")
}
