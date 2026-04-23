package game

import "TheWar/internal/domain/cards"

type PassiveContext struct {
	Trigger             string
	SourcePlayerIndex   int
	ActorPlayerIndex    int
	EventUnitInstanceID string
	AttackTargetID      string
	AttackerInstanceID  string
	DamagedByInstanceID string
	PlayedCardCode      string
	PlayedCardType      string
	PlayedCardIsTank    bool
	PlayedSkillCode     string
	HeroSkillUsed       bool
}

type CardPassiveHandler func(m *MatchState, source *UnitState, owner *PlayerState,
	enemy *PlayerState, spec cards.PassiveSpec, ctx PassiveContext) error

func getCardPassiveHandler(code string) (CardPassiveHandler, bool) {
	handlers := map[string]CardPassiveHandler{
		"coordinated_actions":   HandlePassiveAuraBuff,
		"perfection_of_tactics": HandlePassiveReactiveBuff,
		"doctrinal_burning":     HandlePassiveAuraBuff,
		"barrage":               HandlePassiveReactiveDamage,
		"trench_warfare":        HandlePassiveReactiveBuff,
		"toxic_cloud":           HandlePassiveReactiveDamage,
		"perfect_position":      HandlePassiveReactiveBuff,
		"rapid_reload_protocol": HandlePassiveReactiveBuff,
		"iron_will":             HandlePassiveReactiveBuff,
		"instant_reaction":      HandlePassiveReactiveDamage,
		"protective_field":      HandlePassiveAuraBuff,
		"endless_hate":          HandlePassiveReactiveBuff,
		"to_the_last_drop":      HandlePassiveReactiveBuff,
		"constant_training":     HandlePassiveReactiveBuff,
		"eradication":           HandlePassiveAuraBuff,
		"blood_thirst":          HandlePassiveScalingAuraBuff,
		"lead_by_example":       HandlePassiveReactiveBuff,
		"veteran":               HandlePassiveReactiveBuff,
		"war_machine":           HandlePassiveReactiveBuff,
		"humanity_revenge":      HandlePassiveReactiveDamage,
		"iron_support":          HandlePassiveScalingAuraBuff,
		"many_tons_of_hope":     HandlePassiveReactiveBuff,
		"mechanized_group":      HandlePassiveScalingAuraBuff,
		"countermeasures":       HandlePassiveReactiveDamage,
		"urgent_reinforcement":  HandlePassiveReactiveHeal,
		"single_organism":       HandlePassiveAuraBuff,
		//
		"species_trait":       HandlePassiveReactiveDebuff,
		"horrible_stench":     HandlePassiveReactiveDebuff,
		"shell_evolution":     HandlePassiveReactiveBuff,
		"mass_domination":     HandlePassiveScalingAuraBuff,
		"claw_evolution":      HandlePassiveReactiveBuff,
		"elder_species":       HandlePassiveAuraBuff,
		"new_biomaterial":     HandlePassiveReactiveBuff,
		"poison_burst":        HandlePassiveReactiveDebuff,
		"sharp_spikes":        HandlePassiveReactiveDebuff,
		"aggressive_species":  HandlePassiveScalingAuraBuff,
		"temporary_mutation":  HandlePassiveReactiveBuff,
		"poisonous_presence":  HandlePassiveReactiveDamage,
		"scourge_of_humanity": HandlePassiveScalingAuraBuff,
		"instincts":           HandlePassiveScalingAuraBuff,
		"living_shield":       HandlePassiveAuraBuff,
		"ancient_enemy":       HandlePassiveReactiveDamage,
		"deadliest_enemy":     HandlePassiveReactiveDamage,
		"herald_of_swarm":     HandlePassiveCounterattack,
		"fatal_infection":     HandlePassiveReactiveDebuff,
	}
	h, ok := handlers[code]
	return h, ok
}
