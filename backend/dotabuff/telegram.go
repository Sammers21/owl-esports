package dotabuff

import (
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type TelegramBot struct {
	Engine *Engine
	Token  string
	Bot    *tgbotapi.BotAPI
}

func NewTelegramBot(engine *Engine, token string) *TelegramBot {
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
func (b *TelegramBot) SendPickWinRatesToUser(chatId int64, msgId int, split []string) error {
	reply := msgId != 0
	path, err := b.Engine.GenerateHeatMap(split)
	if err != nil {
		log.Error().Err(err).Msg("Error generating heatmap")
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Error generating heatmap: %v", err))
		if reply {
			msg.ReplyToMessageID = msgId
		}
		b.Bot.Send(msg)
		return err
	}
	log.Info().Msg("Sending heatmap...")
	file, _ := os.Open(path)
	reader := tgbotapi.FileReader{Name: "image.jpg", Reader: file}
	photo := tgbotapi.NewPhoto(chatId, reader)
	if reply {
		photo.ReplyToMessageID = msgId
	}
	photo.Caption = `Here is the counter heatmap of the winrate of the heroes you selected.`
	_, err = b.Bot.Send(photo)
	if err != nil {
		log.Error().Err(err).Msg("Error sending photo")
		return err
	}
	err = os.Remove(path)
	if err != nil {
		log.Error().Err(err).Msg("Error removing file")
		return err
	}
	return nil
}

func (b *TelegramBot) Start() error {
	log.Info().Msg("Starting telegram bot...")
	bot, err := tgbotapi.NewBotAPI(b.Token)
	b.Bot = bot
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
			split := strings.Split(text, ",")
			log.Info().Str("username", update.Message.From.UserName).Str("text", text).Msg("Received message")
			if text == "/start" || text == "/help" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a bot that \n 123123")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			} else if len(split) == 10 {
				b.SendPickWinRatesToUser(update.Message.Chat.ID, update.Message.MessageID, split)
			} else if strings.HasPrefix(text, "https://www.dotabuff.com/matches/") {
				match, err := ExtractHerosFromDBLink(text)
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
