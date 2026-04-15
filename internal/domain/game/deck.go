package game

import (
	"errors"
	"math/rand/v2"

	"github.com/google/uuid"
)

/*А этот файл нужен для того, чтобы определить деку игрока в момент попадания им в игру. И дело тут вот в чем.
Перед матчем игрок обязан составить колоду, притом если ее длина не будет равна 20-игрок не сможет начать матч.
Далее. Эта дека живет в БД, вместе с описанием карты в каждом конкретном экземпляре*/

// данная структура служит описанием деки, которое живет в БД
type DeckEntry struct {
	Kind       string `json:"kind"`        //<-тип карты, либо атакующая либо баф карта
	TemplateID string `json:"template_id"` //<-чертеж карты
	Count      int    `json:"count"`       //<-количество карт в деке
}

// структура служащая для описания карт владения, нужна для сборки матча из колоды, структура которой описана выше
type OwnedCardInfo struct {
	GamerCardID   uint
	Copies        int
	Level         int
	Name          string
	Description   string
	ManaCost      int
	Attack        int
	HealthPoints  int
	CardType      string
	ImageKey      string
	AssetBaseKey  string
	SplashRadius  int
	BaseCooldown  int
	HasSkill      bool
	SkillImageKey string
}

/*
Доменный валидатор колоды. Он определеяет допустима ли колода по правилам игры (не больше ли копий чем можно)
и владения (есть ли у игрока вообще столько копий, сколько он прислал). Так и вот. Принимаем собственно деку,
в этой деке есть как боевые карты, так и бафф карты, а далее, карты владения, тоже как баттл так и баф карты.
Почему карты разделены по темплейтам и владению ? Потому что игрок присылает намерения DeckEntry. А уже овнед
проверяет на ограничение сервера. Так и вот...
*/
func ValidateDeckList(
	entries []DeckEntry,
	battleMax map[string]int,
	buffMax map[string]int,
	ownedBattle map[string]int,
	ownedBuff map[string]int,
) error { //<-если что отдаем ошибку
	total := 0                  //<-предполагаемое количество карт. Здесь я буду дополнять кол-во
	for _, e := range entries { //<-проверяю количество заявленных карт от клиента
		if e.Count <= 0 { //<-и если в этом нет карт, отдаем ошибку
			return ErrDeckCountInvalid
		}
		total += e.Count //<-а так приплюсовываем карты к нашей переменной
		switch e.Kind {  //<-а теперь по типу этих карт проверяем вообще есть ли такие карты
		case "battle":
			max, ok := battleMax[e.TemplateID] //<-смотрим присланный слайс из боевых карт
			if !ok {                           //<-если что то не нравится (нет в природе такой карты)
				return ErrDeckUnknownCard //<-отдаем соответствующую ошибку
			}
			if e.Count > max { //<-если карт больше допустимого значения по копиям одной карты в колоде
				return ErrDeckTooManyCopies
			}
			if ownedBattle[e.TemplateID] < e.Count { //<-если чувак прислал карты, которого у него нет
				return ErrDeckNotOwnedEnough
			}
			//ну и та же система проверок для баф карт
		case "buff":
			max, ok := buffMax[e.TemplateID]
			if !ok {
				return ErrDeckUnknownCard
			}
			if e.Count > max {
				return ErrDeckTooManyCopies
			}
			if ownedBuff[e.TemplateID] < e.Count {
				return ErrDeckNotOwnedEnough
			}
		default:
			return ErrDeckUnknownKind
		}
	}
	//если присланных карт не 20-ошибка
	if total != 20 {
		return ErrDeckSizeNot20
	}
	return nil
}

/*
Доменная функция непосредственного составления деки. Здесь принимаем собственно деку, которую прислал игрок,
и два типа этих карт, после чего мы отдаем ему готовую к игре колоду карт, или ошибку...
*/
func BuildDeck(
	entries []DeckEntry,
	battle map[string]OwnedCardInfo,
	buff map[string]OwnedCardInfo,
) ([]CardsInMatch, error) {
	//создаем массив из 20 карт (слайс блять, я в курсе)
	deck := make([]CardsInMatch, 0, 20)
	//заполняем его епта
	for _, e := range entries {
		//берем инфу о всех картах владения игрока
		var inf OwnedCardInfo
		//понятно
		var ok bool
		//и свичим их по их типу
		switch e.Kind {
		//далее происходит сортировка карт по шаблону
		case "battle":
			inf, ok = battle[e.TemplateID]
		case "buff":
			inf, ok = buff[e.TemplateID]
		default:
			return nil, ErrDeckUnknownKind
		}
		if !ok {
			return nil, ErrDeckUnknownCard
		}
		//а далее уже составляем колоду для матча
		for i := 0; i < e.Count; i++ {
			deck = append(deck, CardsInMatch{
				//присваиваем каждой карте уникальный айдишник
				InstanceID:    NewInstanceID(),
				Kind:          e.Kind,
				TemplateID:    e.TemplateID,
				GamerCardID:   inf.GamerCardID,
				CardLevel:     inf.Level,
				Name:          inf.Name,
				Description:   inf.Description,
				ManaCost:      inf.ManaCost,
				Attack:        inf.Attack,
				HealthPoints:  inf.HealthPoints,
				CardType:      inf.CardType,
				ImageKey:      inf.ImageKey,
				AssetBaseKey:  inf.AssetBaseKey,
				SplashRadius:  inf.SplashRadius,
				BaseCooldown:  inf.BaseCooldown,
				HasSkill:      inf.HasSkill,
				SkillImageKey: inf.SkillImageKey,
			})
		}
	}
	//если карт не 20, то отдаем ошибку
	if len(deck) != 20 {
		return nil, errors.New("internal: buit deck is not 20")
	}
	//перемешиваем деку, чтобы сама по себе колода никогда не была в той последовательности, в которой ее описали
	shuffleDeck(deck)
	return deck, nil
}

/*Таким образом составляется колода, где чувак скидывает колоду, а мы удобно формируем деку
Зачем это вообще нужно ? По сути, функция превращает деклорацию колоды в полноценный рантайм,
где каждой карте я присваиваю уникальный ID и уровнями карты.*/

// Тут интересно. Мы принимаем деку и (псевдо) рандомизируем положение карт внутри массива
func shuffleDeck(deck []CardsInMatch) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// а тут просто функция отдачи айдишника
func NewInstanceID() string {
	return uuid.NewString()
}

/*Таким образом в данном файле собрано все необходимое для того, чтобы успешно как собирать колоду, так и ее
валидировать. Плюс пару хелп функций, которые служат а-чтобы перемешать деку, б-дать каждой карте уникальный
айдишник.*/
