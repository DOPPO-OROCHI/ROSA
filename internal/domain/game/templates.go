package game

type BattleTemplate struct {
	TemplateID    string
	HealthPoints  int
	Attack        int
	SplashRadius  int
	Cooldown      int
	Manacost      int
	IsTank        bool
	CardType      string
	CanBeUpgraded bool
	ImageKey      string
	AssetBaseKey  string
}

type BuffTemplate struct {
	TemplateID   string
	ManaCost     int
	BuffType     string
	BuffValue    int
	OnlyFor      string
	Duration     int
	ImageKey     string
	AssetBaseKey string
}
