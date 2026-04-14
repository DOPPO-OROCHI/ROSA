package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

/*Идемпотентное создание пользователя в БД при его отправке сообщения в бот. Без этого процесс регистрации
не будет реализован никак.*/

func AddNewUser(db *gorm.DB, update tgbotapi.Update) error {
	if update.Message == nil || update.Message.From == nil {
		return nil
	}
	from := update.Message.From
	return EnsureUser(db, Profile{
		TGID:      from.ID,
		Username:  from.UserName,
		FirstName: from.FirstName,
		LastName:  from.LastName,
		Language:  from.LanguageCode,
	})
}
