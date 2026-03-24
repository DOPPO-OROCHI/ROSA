package db

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
)

/*Файл целиком и полностью посвящен функциям, которые по аналогии с заполнением ключей видео/аудио эффектов заполняют
данные на сидинг. Нужны они ровно для этой же цели, только в рамках сидов. Зачем это нужно ? Потому что нельзя писать
image/assets keys руками, ибо в таком случае будет риск ошибок на этапе записи. Плюс ко всему, здесь используются
те же функции, которые используются и для локального заполнения. Таким образом assets_keys отвечает на вопрос, как
строить ключ, а данный файл о том, где и когда применить это же заполнение только для сидинга. Перейдем к коду.*/

/*
Функция для заполнения ключей в рамках целого массива боевых карт
Здесь мы принимаем собственно входящий массив из боевых карт. Так вот...
*/
func FillBattleKeys(ts []cards.BattleCardTemplate) {
	//проходимся циклом по длине всего массива
	for i := range ts {
		//если ключ пуст, заполняем его в соответствии с функцией
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.BattleCardBaseKey(ts[i].CodeString)
		}
		//такая же тема
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
		if ts[i].SkillImageKey == "" {
			ts[i].SkillImageKey = game.SkillImageKey(ts[i].AssetBaseKey)
		}
	}
}

// ровно такая же логика и для баф карт
func FillBuffKeys(ts []cards.BuffCardsTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.BuffCardBaseKey(ts[i].CodeString)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
	}
}

// и для героев
func FillHeroKeys(ts []heroes.CharacterTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.HeroBaseKey(ts[i].CharacterCode)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
		if ts[i].SkillImageKey == "" {
			ts[i].SkillImageKey = game.SkillImageKey(ts[i].AssetBaseKey)
		}
		if ts[i].AttackImageKey == "" {
			ts[i].AttackImageKey = game.AttackImageKeyHero(ts[i].AssetBaseKey)
		}
	}
}

/*Таким образом реализована логика автозаполнения ключей. Приницпиально то, что все происходит в едином
формате, из чего следует то, что разночтений не будет никогда. Я по началу ломался и думал а может просто
использовать функции из assets_keys. Но так не получится сделать, поскольку в таком случае нет возможности
централизированно повлиять на сид для записи в БД, чем и занимается этот файл
*/
