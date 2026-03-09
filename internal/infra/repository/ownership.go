package repository

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/player"

	"gorm.io/gorm"
)

/*Данный файл отвечает владению карт персонажем внутри БД. Именно нижеуказанные структуры задают все то, что хранится у
игрока в БД по его gamerID, в том числе все шаблоны карт, копии бля, уровни и так далее. Когда чувак заходит в бота, он
по сути проходит регистрацию внутри БД, после чего ему выдается базовый набор персонажей + всех карт из стандартного набора.
Здесь и описывается это. Почему это так? Потому что у игрока должна быть возможность прокачивать свои карты (функции прокачки
которой пока нет...), ну и донатики... Перейдем к структурам*/

// Структура владения боевых карт. Все карты привязаны к айди пользователя, который в самом TelegramUser привязан к gormID
// что здесь важно отметить так это именно теги. Честно признаюсь, теги-вайб. Но ниже будет подробный гайд по всем описанным
type GamerBattleCards struct {
	gorm.Model
	GamerID        uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_battle_card"`
	Gamer          player.TelegramUser      `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CardTemplateID uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_battle_card"`
	CardTemplate   cards.BattleCardTemplate `gorm:"foreignKey:CardTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Copies         int                      `gorm:"not null;default:0"`
	CardLevel      int                      `gorm:"not null;default:1"`
	CardXP         int                      `gorm:"not null;default:0"`
}

// По аналогии с боевыми картами, это струтура владения баф картами. Все просто
type GamerBuffCards struct {
	gorm.Model
	GamerID        uint                    `gorm:"not null;index;uniqueIndex:ux_gamer_buff_card"`
	Gamer          player.TelegramUser     `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CardTemplateID uint                    `gorm:"not null;index;uniqueIndex:ux_gamer_buff_card"`
	CardTemplate   cards.BuffCardsTemplate `gorm:"foreignKey:CardTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Copies         int                     `gorm:"not null;default:0"`
	CardLevel      int                     `gorm:"not null;default:1"`
	CardXP         int                     `gorm:"not null;default:0"`
}

/*
функция выгружающая все карты из соответствующих таблиц владения игрока, удобно формируя при этом две мапы,
первая из которых нужна для BuildDeck (когда мы составляем рантайм колоду), а вторая для валидации копий в
ValidateDeckList. Смотри в чем прикол. Когда чувак заходит в бота, он получает стартовый набор карт, да, НО!
Из этих карт надо создать деку. Данная функция вызывается при составлении деки (POST/deck). Серв должен проверить
в натуре ли заявленные карты во владении. Для этого и существуют следующие функции
*/
func LoadOwnedBattleCards(tx *gorm.DB, userID uint) (map[string]game.OwnedCardInfo, map[string]int, error) {
	var rows []GamerBattleCards //<-сюда будем заполнять все данные, которые найдем
	//прогружаем вместе с владением шаблоны карт
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	//создаем мапу из карт, которые мы ранее прогрузили
	info := make(map[string]game.OwnedCardInfo, len(rows))
	//а здесь будет храниться информация о том, сколько копий конкретных карт игрок заявил, и сколькими реально владеет
	copies := make(map[string]int, len(rows))
	for _, r := range rows {
		code := r.CardTemplate.CodeString
		//заполняем всю инфу, превращая все гавно в полноценный рантайм компонент
		info[code] = game.OwnedCardInfo{GamerCardID: r.ID, Copies: r.Copies, Level: r.CardLevel}
		/*Вы можете спросить, а зачем отдельно брать мапу из копий? Вопрос резонный и честно я тупанул, но уже поздно
		что то менять. Вторая мапа служит для быстрого входа в ValidateDeckList, где проверяются конкретно копии. Да,
		окей, это тупо, потому что можно было просто передать r.Copies, но делать уже нехер.*/
		copies[code] = r.Copies
	}
	return info, copies, nil
}

// та же самая функция только под баф карты
func LoadOwnedBuff(tx *gorm.DB, userID uint) (map[string]game.OwnedCardInfo, map[string]int, error) {
	var rows []GamerBuffCards
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	info := make(map[string]game.OwnedCardInfo, len(rows))
	copies := make(map[string]int, len(rows))
	for _, r := range rows {
		code := r.CardTemplate.CodeString
		info[code] = game.OwnedCardInfo{GamerCardID: r.ID, Copies: r.Copies, Level: r.CardLevel}
		copies[code] = r.Copies
	}
	return info, copies, nil
}

/*Я понимаю что слишком много движух вокруг деки, да. Но блин, прикол в том что тут же и реализованы все слои,
которые ну тупо необходимы для того, чтобы все это добро валидировать, притом на разных фазах, начиная от БД,
заканчивая DTO. Так и вот, данный файл чисто посвящен тому, чтобы валидировать все овнер карты игрока для составления
деки. Отсюда мы выгружаем всю необходимую об этом инфу, которая и используется в билдах а так же валидаторе. Вторая
мапа да, кринж, но блин, тупанул.*/
