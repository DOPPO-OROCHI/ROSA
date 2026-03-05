package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/infra/repository"
	"net/http"

	"gorm.io/gorm"
)

/*Плавно переходим от деки, карт и так далее к героям. Собственно этот файл целиком и полностью об этом
А вообще круто что есть такое разделение жесткое. Ну сам рассуди, вот я захочу операции над героями добавить
для пользователя. Единственное что мне придется сделать -> добавить эту операцию. Удобно я расположил файлы,
не зря 3 раза весь код пересобирал после потери...*/

// классическая структура зависимостей
type HeroListHandler struct {
	DB *gorm.DB
}

// хендлер для получения всех героев во владении игрока
func NewHeroesListHandler(d HeroListHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-классическая преамбула
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID                   //<-записываем пользователя в нашу память, для операционки
		var owned []repository.GamerCharacter //<-а сюда я буду сохранять всех героев из БД игрока
		/*основная причина ебки и хейта себя и собственных возможностей. Короче Preload это умная штука вася. Дело
		в том, что помимо владения игроком, я ж еще должен подгрузить связанные сущности (поскольку владение учитывает
		лишь уровень перса, его прогресс и собственно айди игрока, ни о каких статах там речи не идет). Зачем я так
		сделал вообще ? Потому что уровни, абилки, статы и все такое исходит из базовых значений. Представляю мое ебало
		если бы я так НЕ сделал, когда придется балансить всю муру. Так и вот. Вместе с нахождением в БД именно персов
		владения я дополнительно должен подргрузить еще и чертежи, потому что сама по себе инфа о владении игроку не скажет
		ровным счетом ничего (по крайней мере того, что он реально хочет увидеть). Круто.*/
		if err := d.DB.Preload("CharacterTemplate").Where("gamer_id = ?", userID).Find(&owned).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		//а потом я все это гавно формирую в удобоваримый слайс, где уже и заполняю его
		out := make([]dto.OwnedHeroDTO, 0, len(owned))
		for _, g := range owned {
			t := g.CharacterTemplate //<-важная вещь, потому что как уже говорилось, владение не отражает реальные статы
			//реальные статы отражает немплейт бля.
			out = append(out, dto.OwnedHeroDTO{
				HeroID:         t.ID,
				HeroCode:       t.CharacterCode,
				Name:           t.Name,
				Level:          g.CharacterLevel,
				HealthPoints:   t.HealthPoints,
				AttackPower:    t.AttackPower,
				AttackCooldown: t.AttackCooldown,
				SplashRadius:   t.SplashRadius,
				Description:    t.Description,
				ImageKey:       t.ImageKey,
				AssetBaseKey:   t.AssetBaseKey,
			})
		}
		//отдаем челику ответ на его запрос
		middleware.WriteJSON(w, http.StatusOK, dto.HeroesListResponce{Heroes: out})
	}
}

/*Таким образом получается следующий кейс. Как обычно мы принимаем пользовательский запрос, да, проверяем пользователя
на аутентификацию, все круто, а потом собственно занимаемся логикой, в которую входит по большей части составление ответа.
Главный гемор в этом хендлере Preload, который нужен чтобы подгрузить связанные сущности, поскольку владение не дает никакой
важной инфы. Ну и после этого составляем собственно DTOшку для отдачи пользователю, серриализируя созданный раннее слайс в
JSON. Крутой ХЕНДЛЕР!*/
