package game

import "strings"

func cleanKeyPart(s string) string {
	s = strings.TrimSpace(s)
	return strings.Trim(s, "/")
}

func buildKey(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		p = cleanKeyPart(p)
		if p == "" {
			continue
		}
		cleaned = append(cleaned, p)
	}
	return strings.Join(cleaned, "/")
}

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
