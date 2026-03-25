package game

import (
	"TheWar/internal/domain/cards"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type PassiveSkillHandler func(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error

type PassiveTriggerContext struct {
	AttackerInstanceID string
	AttackerOwnerIdx   int

	TargetInstanceID string
	DeadInstanceID   string
	SourceOwnerIdx   int
	TargetOwnerIdx   int
	TargetSlot       int
}

type passivePrepared struct {
	targets []*UnitState
	value   int
}

var PassiveSkillsHandler map[string]PassiveSkillHandler

func init() {
	PassiveSkillsHandler = map[string]PassiveSkillHandler{
		"disgusting_stench": passiveApplyEffectOnHitMe,       // МРАЗЬ
		"predatory_beast":   passiveGenericSkillCooldownDown, // ДЕМОНИЧЕСКИЙ БЕРСЕРК
	}
}

/*
Диспетчер одной пассивки одной конкретной карты. В чем прикол ?
Здесь мы принимаем вообще все что связано с картой, проверяя ее
триггер, после чего ищем его хендлер (код пассивки->хендлер), после
чего вызываем его. Такой удобный мини оркестратор пассивок. Круто.
*/
func triggerPassiveByTrigger(m *MatchState,
	ownerIdx int, source *UnitState,
	trigger string, ctx PassiveTriggerContext) error {
	if m == nil || source == nil {
		return nil
	}
	if !shouldTriggerPassive(source, trigger) {
		return nil
	}
	h, err := getPassiveSkillHandler(source.PassiveCode)
	if err != nil {
		return err
	}
	return h(m, ownerIdx, source, trigger, ctx)
}

/*
-------ХЕНДЛЕРЫ ПАССИВНЫХ СКИЛЛОВ-------
Здесь описываются собственно чертежи пассивных умений, к которым
будут подключены пассивки каждой отдельной карты. Здесь проверяются
параметры карты
*/
func passiveGenericDamageUp(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveAttackBuff(prep.targets, prep.value)
	return nil
}

func passiveGenericHPUp(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveHPBuff(prep.targets, prep.value)
	return nil
}

func passiveGenericCooldownDown(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveAttackCooldownDown(prep.targets, prep.value)
	return nil
}

func passiveGenericSkillDamageUp(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveSkillDamageBuff(prep.targets, prep.value)
	return nil
}

func passiveGenericSkillCooldownDown(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveSkillCooldownDown(prep.targets, prep.value)
	return nil
}

func passiveGenericHeal(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveHeal(prep.targets, prep.value)
	return nil
}

func passiveGenericDamage(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok || prep.value <= 0 {
		return nil
	}
	return applyPassiveDamage(m, ownerIdx, prep.targets, prep.value)
}

// НАЛОЖЕНИЕ ДЕБАФА НА ПРОТИВНИКА, КОТОРЫЙ БЬЕТ КАРТУ-ИСТОЧНИК
func passiveApplyEffectOnHitMe(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	prep, ok := preparePassiveBase(m, ownerIdx, source, event, ctx)
	if !ok {
		return nil
	}
	applyPassiveEffect(prep.targets, source.PassiveEffect, source.PassiveDuration, prep.value)
	return nil
}

/*
-------ХЕЛПЕРЫ-------
*/

/*
Общий подготовительный хелпер для пассивки. Здесь проводятся базовые проверки, такие как входящие аргументы
например... Ну ладно-ладно, здесь проверяем вообзще должна ли карта играть сейчас, проверяется спец-случай,
онхитми, который отвечает за то, что пиздят карту и является ли это поводом для тика какой-либо пассивки. При
этом отдаем структуру, которая в себе несет цели для пассивки и значение какого либо из тиков.
*/
func preparePassiveBase(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) (passivePrepared, bool) {
	if m == nil || source == nil || ownerIdx < 0 || ownerIdx > 1 {
		return passivePrepared{}, false
	}
	if !shouldTriggerPassive(source, event) {
		return passivePrepared{}, false
	}
	if source.PassiveTrigger == cards.PassiveTriggerHitMe && ctx.TargetInstanceID != source.InstanceID {
		return passivePrepared{}, false
	}
	matched := passiveMatchedCount(m, ownerIdx, source)
	if !passiveConditionOK(m, ownerIdx, source, matched) {
		return passivePrepared{}, false
	}
	value := calcPassiveFinalValue(source, matched)
	if value == 0 {
		return passivePrepared{}, false
	}
	targets := passiveTargets(m, ownerIdx, source, ctx)
	if len(targets) == 0 {
		return passivePrepared{}, false
	}
	return passivePrepared{targets: targets, value: value}, true
}

/*
Считаем сколько карт подходит под условие пассивки. Дело в том, что для некоторых эффектов необходимо
присутствие одного или нескольких карт того, или иного типа. Все это формируется здесь. Отдаем int,
чтобы дальше использовать это в passiveConditionOK, где будет сопоставляться количественные требования
для условий пассивки. Типа, если на столе будет 5 демонов -хопаааа... То есть че, сюда попадает карта, у
которой есть фильтр пассивки PassiveCountType. Эта функция собирает все, что подходит под эту штуку. Грубо
говоря тут выбираем кого считать для активации пассивки. Не просто нихрена.
*/
func passiveMatchedCount(m *MatchState, ownerIdx int, source *UnitState) int {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return 0
	}
	ally := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	countInTable := func(t [TableSize]*UnitState) int {
		c := 0
		for i := 0; i < TableSize; i++ {
			u := t[i]
			if u == nil {
				continue
			}
			if source.PassiveCountType != "" {
				if u.CardType == source.PassiveCountType {
					c++
				}
				continue
			}
			if source.PassiveCountCode != "" && u.TemplateID == source.PassiveCountCode {
				c++
			}
		}
		return c
	}
	switch source.PassiveCountOwner {
	case cards.PassiveCountOwnerEnemy:
		if enemy == nil {
			return 0
		}
		return countInTable(enemy.Table)
	case cards.PassiveCountOwnerBoth:
		total := 0
		if ally != nil {
			total += countInTable(ally.Table)
		}
		if enemy != nil {
			total += countInTable(enemy.Table)
		}
		return total
	default:
		if ally == nil {
			return 0
		}
		return countInTable(ally.Table)
	}
}

/*
Проверяем активна ли пассивка исходя из текущего состояния матча.
Сюда передаем карту-источник пассивки, и уже посчитанное функцией
выше количество совпадающих карт с требованиями. Дальше смотрим на
PassiveCondition и решаем, может ли считаться пассивка включенной.
*/
func passiveConditionOK(m *MatchState, ownerIdx int, source *UnitState, matched int) bool {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return false
	}
	switch source.PassiveCondition {
	//пассивка работает всегда (если пустое поле в дефолтах-работает всегда)
	case "", cards.PassiveConditionAlways:
		return true
		//если найдено не меньше Х карт
	case cards.PassiveConditionCountAtLeast:
		return matched >= source.PassiveConditionCount
		//не больше Х карт
	case cards.PassiveConditionCountAtMost:
		return matched <= source.PassiveConditionCount
		//ровно Х карт
	case cards.PassiveConditionExact:
		return matched == source.PassiveConditionCount
		//если на столе карта нужного типа
	case cards.PassiveConditionDemonicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.DemonicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionMechanicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.MechanicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionOrganicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.OrganicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionHealerOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.HealerCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
		//после всей проверки отдаем инфу о том, считается ли пассивка активной, или еще нет
	default:
		return false
	}
}

/*
Собираем список целей для пассивки. И дело тут вот в чем. Каждая из типов пассивок действует по разному. Кто то
ебашим по всему столу вообще, кто то только по вражескому столу, кто то случайно, кто то вообще по своим и так
далее, список огромный. Так вот, здесь все это собирается в единый массив, который будет отдаваться в хендлеры,
для применения. Если таргет не задан, функция берет цель из контекста, через ctx
*/
func passiveTargets(m *MatchState, ownerIdx int, source *UnitState, ctx PassiveTriggerContext) []*UnitState {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return nil
	}
	ally := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	if ally == nil || enemy == nil {
		return nil
	}
	out := make([]*UnitState, 0, 6)
	add := func(u *UnitState) {
		if u != nil {
			out = append(out, u)
		}
	}
	addAll := func(t [TableSize]*UnitState) {
		for i := 0; i < TableSize; i++ {
			add(t[i])
		}
	}
	addByType := func(t [TableSize]*UnitState, cardType string) {
		for i := 0; i < TableSize; i++ {
			u := t[i]
			if u != nil && u.CardType == cardType {
				out = append(out, u)
			}
		}
	}
	switch source.PassiveTarget {
	case cards.PassiveTargetSelf:
		add(source)
	case cards.PassiveTargetAllyAll:
		addAll(ally.Table)
	case cards.PassiveTargetEnemyAll:
		addAll(enemy.Table)
	case cards.PassiveTargetBothAll:
		addAll(ally.Table)
		addAll(enemy.Table)
	case cards.PassiveTargetAllyLeftRight:
		slot, _ := ally.FindSlot(source.InstanceID)
		if slot-1 >= 0 {
			add(ally.Table[slot-1])
		}
		if slot+1 < TableSize {
			add(ally.Table[slot+1])
		}
	case cards.PassiveTargetRandomEnemy:
		var picked *UnitState
		seen := 0
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			seen++
			if rng.Intn(seen) == 0 {
				picked = u
			}
		}
		add(picked)
	case cards.PassiveTargetRandomAlly:
		var picked *UnitState
		seen := 0
		for i := 0; i < TableSize; i++ {
			u := ally.Table[i]
			if u == nil {
				continue
			}
			seen++
			if rng.Intn(seen) == 0 {
				picked = u
			}
		}
		add(picked)
	case cards.PassiveTargetAttacker:
		if ctx.AttackerOwnerIdx < 0 || ctx.AttackerOwnerIdx > 1 || ctx.AttackerInstanceID == "" {
			break
		}
		p := m.Players[ctx.AttackerOwnerIdx]
		if p == nil {
			break
		}
		_, attacker := p.FindSlot(ctx.AttackerInstanceID)
		add(attacker)
	case cards.PassiveTargetAllyTypeDemonical:
		addByType(ally.Table, cards.DemonicalCard)
	case cards.PassiveTargetAllyTypeMechanical:
		addByType(ally.Table, cards.MechanicalCard)
	case cards.PassiveTargetAllyTypeOrganical:
		addByType(ally.Table, cards.OrganicalCard)
	case cards.PassiveTargetAllyTypeHealer:
		addByType(ally.Table, cards.HealerCard)
	case cards.PassiveTargetEnemyTypeDemonical:
		addByType(enemy.Table, cards.DemonicalCard)
	case cards.PassiveTargetEnemyTypeMechanical:
		addByType(enemy.Table, cards.MechanicalCard)
	case cards.PassiveTargetEnemyTypeOrganical:
		addByType(enemy.Table, cards.OrganicalCard)
	case cards.PassiveTargetEnemyTypeHealer:
		addByType(enemy.Table, cards.HealerCard)
	default:
		if ctx.TargetOwnerIdx == ownerIdx && ctx.TargetSlot >= 0 && ctx.TargetSlot < TableSize {
			add(ally.Table[ctx.TargetSlot])
		} else if ctx.TargetOwnerIdx == 1-ownerIdx && ctx.TargetSlot >= 0 && ctx.TargetSlot < TableSize {
			add(enemy.Table[ctx.TargetSlot])
		}
	}
	return out
}

/*
Считаем финальное значение пассивки. Здесь принимаем карту и кол-во совпаденеий, которое
было найдено ранее. Функция нужна для того, чтобы узнать итоговую силу эффекта, которая
возвращает либо текущее значение, либо перемножает его на количество совпадений. На выходе
даем готовое число, которое двиглом и мной интерпретируется как сила пассивки.
*/
func calcPassiveFinalValue(source *UnitState, matchedCount int) int {
	if source == nil {
		return 0
	}
	switch source.PassiveScale {
	case cards.PassiveScalePerCount:
		return source.PassiveValue * matchedCount
	case cards.PassiveScaleFlat:
		fallthrough
	default:
		return source.PassiveValue
	}
}

/*
-------ПРИМЕНЯЕМ КАКОЙ ЛИБО ПАРАМЕТР-------
Далее по сути лишь инструменты, которые легко можно интегрировать в любой из
хендлеров. Смысл прост, здесь раскидано все то, что необходимо к применению.
*/

// БАФФ НА АТАКУ
func applyPassiveAttackBuff(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.Attack += value
	}
}

// БАФФ НА ПОДНЯТИЕ ХП И ХИЛЛ
func applyPassiveHPBuff(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.HP += value
		t.MaxHP += value
	}
}

// ХИЛИМ, БЕЗ ПОВЫШЕНИЯ МАКСИМАЛЬНОГО ХП
func applyPassiveHeal(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.HP += value
		if t.HP > t.MaxHP {
			t.HP = t.MaxHP
		}
	}
}

// БАФФ НА СНИЖЕНИЕ КД ОСНОВНОЙ АТАКИ КАРТЫ
func applyPassiveAttackCooldownDown(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.Cooldown -= value
		if t.Cooldown < 0 {
			t.Cooldown = 0
		}
	}
}

// БАФФ НА СИЛУ СКИЛА КАРТЫ
func applyPassiveSkillDamageBuff(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.SkillValue += value
	}
}

// БАФФ НА ОТКАТ КД СКИЛА КАРТЫ
func applyPassiveSkillCooldownDown(targets []*UnitState, value int) {
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.SkillCooldownLeft -= value
		if t.SkillCooldownLeft < 0 {
			t.SkillCooldownLeft = 0
		}
	}
}

// БАФФ
func applyPassiveDamage(m *MatchState, ownerIdx int, targets []*UnitState, value int) error {
	if m == nil || ownerIdx < 0 || ownerIdx > 1 {
		return nil
	}
	dead := map[string]struct{}{}
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.HP -= value
		if t.HP <= 0 {
			t.HP = 0
			dead[t.InstanceID] = struct{}{}
		}
	}
	for id := range dead {
		if p := m.Players[ownerIdx]; p != nil {
			if slot, u := p.FindSlot(id); slot >= 0 && u != nil && u.HP <= 0 {
				if err := killUnitAt(m, ownerIdx, slot); err != nil {
					return err
				}
				continue
			}
		}
		enemyIdx := 1 - ownerIdx
		if p := m.Players[enemyIdx]; p != nil {
			if slot, u := p.FindSlot(id); slot >= 0 && u != nil && u.HP <= 0 {
				if err := killUnitAt(m, enemyIdx, slot); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

/*
Используем для наложения временного эффекта на карту\ы. Принимает таргеты,
тип эффекта, длительность и значение, после чего вызываем AddEfect, накидывая
всю хуйню на карту\ы
*/
func applyPassiveEffect(targets []*UnitState,
	effectType string, duration int, value int) {
	for _, target := range targets {
		if target == nil {
			continue
		}
		AddEffect(target, UnitEffect{
			EffectType: effectType,
			TurnsLeft:  duration,
			Value:      value,
		})
	}
}

/*
-------ОБЩИЕ ХЕЛПЕРЫ-------
Здесь идет базовая проверка того, когда должен примениться триггер
пассивки. Плюс функция поиска из карты со всеми пассивными эффектами
по коду карты, который вшит в карту и должен совпадать с defaults.go
*/
func shouldTriggerPassive(u *UnitState, event string) bool {
	return u != nil && u.PassiveCode != "" && u.PassiveTrigger == event
}

func getPassiveSkillHandler(code string) (PassiveSkillHandler, error) {
	h, ok := PassiveSkillsHandler[code]
	if !ok {
		return nil, ErrCardPassiveSkillUnsupported
	}
	return h, nil
}
