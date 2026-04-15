package main

import (
	"TheWar/adapters/telegram"
	"TheWar/internal/infra/db"
	"context"
	"errors"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramWebAppInfo struct {
	URL string `json:"url"`
}

type telegramWebAppButton struct {
	Text   string              `json:"text"`
	WebApp *telegramWebAppInfo `json:"web_app,omitempty"`
}

type telegramReplyKeyboardMarkup struct {
	Keyboard        [][]telegramWebAppButton `json:"keyboard"`
	ResizeKeyboard  bool                     `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool                     `json:"one_time_keyboard,omitempty"`
}

func runTelegramBot(ctx context.Context) error {
	api := db.GoDotEnvVariable("BOT_API")
	if api == "" {
		return errors.New("nil telegram bot api")
	}
	webAppURL := db.GoDotEnvVariable("WEBAPP_URL")
	if webAppURL == "" {
		return errors.New("nil URL")
	}
	bot, err := tgbotapi.NewBotAPI(api)
	if err != nil {
		return err
	}
	bot.Debug = false
	log.Printf("telegram bot connected: %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			log.Printf("telegram bot shutting down")
			return nil
		case update, ok := <-updates:
			if !ok {
				return errors.New("telegram update ch closed")
			}
			if update.Message == nil {
				continue
			}
			errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "An error occurred while registering. Please try your service request again later.")
			if err := telegram.AddNewUser(db.DB, update); err != nil {
				log.Printf("add user error: %v", err)
				if _, err := bot.Send(errMsg); err != nil {
					log.Printf("err while send errMsg : %v", err)
				}
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to service soldier. Begin combat against the enemy immediately. May the Holy Land aid you.")
			msg.ReplyMarkup = telegramReplyKeyboardMarkup{
				ResizeKeyboard: true,
				Keyboard: [][]telegramWebAppButton{
					{
						{
							Text: "Play",
							WebApp: &telegramWebAppInfo{
								URL: webAppURL,
							},
						},
					},
				},
			}
			_, _ = bot.Send(msg)
		}
	}
}
