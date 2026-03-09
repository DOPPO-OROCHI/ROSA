package repository

import (
	"TheWar/internal/domain/cards"

	"gorm.io/gorm"
)

/*Файл посвящен выгрузке глобальных лимитов из шаблонов. Вот в чем мем. У основных шаблонов есть лимиты да...
Но как рантайм об этом узнает ? Без этой функции пришлось бы тупо верить пользователю относительно лимитов.
А так нет, нихера, у нас есть функция которая четко отдает инфу о карте (по коду) и количеству лимита (инт).
Зачем нужно и где используется? Дак в том же валидаторе, он то должен узнать о том, сколько карт должно быть,
так же в формировании деки в CreateMatchTX*/

// Так и че. Принимаем необходимые зависимости (БД переменную), отдаем две мапы с перечислениями лимитов. Круто. Ключ-код карты
func LoadTemplateLimits(tx *gorm.DB) (battleMax map[string]int, buffMax map[string]int, err error) {
	//сюда будем батлы записывать
	var battleTpl []cards.BattleCardTemplate
	//селект тут чисто чтобы именно эти поля взять и никакие другие
	if err := tx.Select("code_string", "max_copies").Find(&battleTpl).Error; err != nil {
		return nil, nil, err
	}
	//сюда заполняем всю инфу по коду
	battleMax = make(map[string]int, len(battleTpl))
	for _, t := range battleTpl {
		battleMax[t.CodeString] = t.MaxCopies
	}
	//та же хуйня для баф карт
	var buffTpl []cards.BuffCardsTemplate
	if err := tx.Select("code_string", "max_copies").Find(&buffTpl).Error; err != nil {
		return nil, nil, err
	}
	buffMax = make(map[string]int, len(buffTpl))
	for _, t := range buffTpl {
		buffMax[t.CodeString] = t.MaxCopies
	}
	//возвращаем инфу
	return battleMax, buffMax, nil
}

/*Такая реализация позволяет абсолютно точно достать все глобальные лимиты карт внутрь памяти. Прикольно.
Нуно чтобы не проебаться и шарить точно за то, сколько карт с собой может взять игрок*/
