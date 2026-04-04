package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/infra/repository"
	"net/http"

	"gorm.io/gorm"
)

/*Данный хендлер отвечает за то, чтобы отдать игроку список его карт. Ничего сложного, только чтение.
Для реализации этой идеии нам, по примеру с ApplyAction хендлером понадобится структура, которая будет
отвечать за зависимости. В нашем случае тут ничего сложного, мы просто принимаем БД переменную, поскольку
наш кейс заключается чисто в чтении данных из БД с их последующей сериализацией. НО. Раз уж все так
легко и все такое, почему я использую структуру? А потому что так я показываю какие у меня крутые яйца в
вопросах масштабируемости. Дело тут вот в чем. В будущем здесь могут быть добавлены еще миллион аргументов,
каждый из которых будет так или иначе использоваться. Чтобы код не был помойкой, было принято решение использовать
структуры, куда в будущем я смогу класть те зависимости, которые я хочу.*/

type CardListHandlerDeps struct {
	DB *gorm.DB
}

/*
Короче с практикой использования структур в качестве аргументов с зависимостями думаю понятно и доходчиво.
Теперь перейдем к описанию хендлера. Как уже сверх логично и понятно, здесь мы отдаем хендлер... Круто да ???
*/
func NewCardsListHandler(d CardListHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-проверяем аторизацию
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID                                                  //<-пишем юзера в память для операционки
		battleRows, err := repository.LoadOwnedBattleCardsRows(d.DB, userID) //<-выгружаем все карты игрока в память
		if err != nil {                                                      //<-коль что то пошло по пизде -> сервер отлетел
			middleware.WriteErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		buffRows, err := repository.LoadOwnedBuffCardsRows(d.DB, userID) //<-точно так же и здесь
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out := dto.CardsListResponse{ //<-а здесь формируем карты игрока, заботливо расфосовывая их по массивам
			//для того, чтобы передать эти карты в ответе
			Battle: make([]dto.OwnedBattleCardsDTO, 0, len(battleRows)),
			Buff:   make([]dto.OwnedBuffCardsDTO, 0, len(buffRows)),
		}
		//а вот это супер важная хрень, над которой я потел
		//здесь мы в цикле заполняем наши слайсы, да, но тут важно то, что по факту есть две стороны одной медали.
		//r-владение игрока, те статы, которые мы будем скейлить, в то время как t- темплейт, чертеж карты, от которой
		//и будут происходить все скейлы. Поэтому нужно было вынести темлейт в отдельную переменную
		for _, r := range battleRows {
			t := r.Tpl
			var skill *dto.BattleSkillDTO
			if t.HasSkill {
				skill = &dto.BattleSkillDTO{
					Code:         t.SkillCode,
					Kind:         t.SkillKind,
					Targeting:    t.SkillTargeting,
					Power:        t.SkillPower,
					BaseCooldown: t.SkillBaseCooldown,
					Duration:     t.SkillDuration,
					ExtraValue:   t.SkillExtraValue,
					IgnoreTank:   t.SkillIgnoreTank,
					HitCount:     t.SkillApplyCount,
				}
			}
			out.Battle = append(out.Battle, dto.OwnedBattleCardsDTO{
				Kind:          dto.CardKindBattle,
				TemplateID:    t.CodeString,
				Name:          t.Name,
				Description:   t.Description,
				CardType:      t.CardType,
				ManaCost:      t.ManaCost,
				HealthPoints:  t.HealthPoints,
				Attack:        t.Attack,
				SplashRadius:  t.SplashRadius,
				BaseCooldown:  t.BaseCooldown,
				IsTank:        t.IsTank,
				MaxCopies:     t.MaxCopies,
				OwnedCardID:   r.OwnedID,
				Copies:        r.Copies,
				Level:         r.Level,
				XP:            r.XP,
				ImageKey:      t.ImageKey,
				AssetBaseKey:  t.AssetBaseKey,
				SkillImageKey: t.SkillImageKey,
				HasSkill:      t.HasSkill,
				Skill:         skill,
			})
		}
		for _, r := range buffRows {
			t := r.Tpl
			out.Buff = append(out.Buff, dto.OwnedBuffCardsDTO{
				Kind:         dto.CardKindBuff,
				TemplateID:   t.CodeString,
				Name:         t.Name,
				Description:  t.Description,
				ManaCost:     t.ManaCost,
				BuffType:     t.BuffType,
				BuffValue:    t.BuffValue,
				OnlyFor:      t.OnlyFor,
				Duration:     t.Duration,
				MaxCopies:    t.MaxCopies,
				OwnedCardID:  r.OwnedID,
				Copies:       r.Copies,
				Level:        r.Level,
				XP:           r.XP,
				ImageKey:     t.ImageKey,
				AssetBaseKey: t.AssetBaseKey,
			})
		}
		//после успешного заполнения всей муры, отдаем пользователю карты, в которых уже есть все необходимое
		middleware.WriteJSON(w, http.StatusOK, out)
	}
}

/*Важно понимать про карты следующее (позднее я еще напишу об этом). Уровень карты напрямую (в интерфейсе игрока) никак
(пока что) не отражается на характеристиках а считается непосредственно в матче. Согласен, пока что это сыро и тупо, в
будущем я сделаю отдельный функицонал под левелинг, чтобы он менялся в БД непосредственно, во владении, но пока имеем что
имеем.*/
