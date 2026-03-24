package game

type PassiveSkillHandler func(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error

type PassiveTriggerContext struct {
	ActorInstanceID  string
	TargetInstanceID string
	DeadInstanceID   string
}

var PassiveSkillsHandler map[string]PassiveSkillHandler

func init() {
	PassiveSkillsHandler = map[string]PassiveSkillHandler{}
}

func getPassiveSkillHandler(code string) (PassiveSkillHandler, error) {
	h, ok := PassiveSkillsHandler[code]
	if !ok {
		return nil, ErrCardPassiveSkillUnsupported
	}
	return h, nil
}
