package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type GetMeHandler struct {
	DB *gorm.DB
}

/*
Хендлер для предоставления информации об игроке, его имени, айди, юзернейме и прочее. Пока что он не учитывает
аватарку и прочие медиаданные.
*/
func NewGetMeHandler(d GetMeHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //<-проверяем авториазацию
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second) //<-добавляем таймаут чтобы иметь возможность отменить действие
		//к примеру, когда ответ слишком долгий, или у чувака плохое соединение
		defer cancel()
		q := d.DB.WithContext(ctx)                           //<-присваиваем переменной нашу глобальную переменную с контекстом
		var u player.TelegramUser                            //<-объявляем пользователя в память
		if err := q.First(&u, au.UserID).Error; err != nil { //<-достаем пользователя из БД
			middleware.WriteErr(w, http.StatusNotFound, "user not found")
			return
		}
		resp := dto.MeResponse{ //<-и записываем пользователя в DTO
			UserID:               u.ID,
			TGID:                 u.TGID,
			Username:             u.Username,
			FirstName:            u.FirstName,
			Rating:               u.Rating,
			XP:                   u.XP,
			SelectedHeroTemplate: u.SelectedHeroTemplateID,
		}
		if u.SelectedHeroTemplateID != nil { //<-составляем персонажа, если вообще выбран (ибо может быть не выбран вообще)
			var tpl heroes.CharacterTemplate
			if err := q.Select("id", "character_code", "name").First(&tpl, *u.SelectedHeroTemplateID).Error; err == nil {
				resp.SelectedHeroCode = tpl.CharacterCode //<-записываем необходимую информацию
				resp.SelectedHeroName = tpl.Name
			}
		}
		middleware.WriteJSON(w, http.StatusOK, resp) //<-и отдаем все, что есть в БД на пользователя
	}
}

/*Таким образом, мы получаем хендлер, который отвечает за отдачу всю информации о пользователе, которая у нас есть. Пока что в коде нет
хендлеров, которые позволяют проводить операции над изменением полей внутри такой штуки как пользователь (никнейм, к примеру), но это
все будет добавляться здесь же, поскольку принцип остается точно таким же, за исключением той детали, что в таком случае останется отдать
только статус код, который скажет либо об успехе, либо об ошибке. Важное что здесь есть -добавление контекста. Это нужная вещь, чтобы не
держать запрос вечно. Позже я добавлю контекст везде. Важно здесь то, что мы не сразу добавить инфу о персонаже, поскольку он может быть
не выбран. Поэтому я и добавляю персонажа отельно и только в том случае, если он выбран. Все*/
