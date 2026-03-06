package game

import "strings"

/*Файл целиком и полностью посвящен функциям-хелперам, которые помогают в заполнении в основном
ключей ассетов для VFX и SFX. Почему это нужно ? Да потому, что так я привожу все заполнение к
единому страндарту, чтобы не было разночтений. Приступим к описанию функций.*/

/*Данная функция принимает в себя строку, отдавая при этом тоже строку,
но уже в отформатированном виде. Замут в чем ? А в том, чтобы удалить все
лишнее и не допустить ошибок, которые заруинят заполнение строк, убирая
все лишние слеши, пробелы и прочий мусор*/
func cleanKeyPart(s string) string {
	s = strings.TrimSpace(s)
	return strings.Trim(s, "/")
}

/*А эта функция как раз таки и формирует чистый path, который в дальнейшем искользуется
для построения ключей. Воооот... Принимаем множества строк, типа "cards", "/battle" и
прочее, а возвращает прикольный и отформатированный текст типа -"cards/battle/и так далее"*/
func buildKey(parts ...string) string {
	//составляем слайс стрингов с емкостью входящих аргументов
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		//заполняем этот слайс
		p = cleanKeyPart(p)
		//если строка пустая, продолжаем цикл, похер
		if p == "" {
			continue
		}
		//отдаем результаты в отдаваемый слайс
		cleaned = append(cleaned, p)
	}
	//отдаем весь слайс как одну строку, с добавлением между элементами слеша
	return strings.Join(cleaned, "/")
}

/*А в этой, как и в последующих чисто функции для написания все тех же ключей,
только уже под каждую сущность. С помощью всех вышеописанных замутов я и реализовываю
чистоут написания. Что круто и оч вайбово!*/
func BattleCardBaseKey(code string) string {
	return buildKey("cards", "battle", code)
}

func BuffCardBaseKey(code string) string {
	return buildKey("cards", "buff", code)
}

func HeroBaseKey(code string) string {
	return buildKey("heroes", code)
}

func ImageKey(base string) string {
	return buildKey(base, "image")
}

func BuildVFXKey(base, action string) string {
	return buildKey(base, "vfx", action)
}

func BuildSFXKey(base, action string) string {
	return buildKey(base, "sfx", action)
}

/*Смотри в чем прикол. Я бы мог просто написать функции, которые уже заранее ставят везде
слеш правильно. НО! Так я рискую со временем проебать единый формат. Так что все вышеописанные
приколы (помимо сложности написания этих функций, хотя, больше логики конструкции) служат для
стабильной конструкции. А так, вот, считай по ГОСТу заполнение*/
