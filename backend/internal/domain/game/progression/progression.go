package progression

import "time"

var levelThresholds = []int{100, 250, 450, 700, 1000, 1350}

func LevelForXP(xp int) (level, next int) {
	level = 1
	for _, threshold := range levelThresholds {
		if xp < threshold {
			return level, threshold
		}
		level++
	}
	return level, levelThresholds[len(levelThresholds)-1] + 350
}

func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	index := level - 2
	if index >= len(levelThresholds) {
		return levelThresholds[len(levelThresholds)-1] + 350
	}
	return levelThresholds[index]
}

func StageForLevel(level int) string {
	switch {
	case level >= 7:
		return "adult"
	case level >= 4:
		return "teen"
	case level >= 2:
		return "child"
	default:
		return "egg"
	}
}

func StatusForBattery(battery int) string {
	switch {
	case battery <= 0:
		return "exhausted"
	case battery <= 25:
		return "tired"
	case battery <= 60:
		return "normal"
	default:
		return "happy"
	}
}

var moscowLocation = loadMoscow()

func MoscowDate(t time.Time) string { return t.In(moscowLocation).Format("2006-01-02") }
func MoscowMidnight(t time.Time) time.Time {
	local := t.In(moscowLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, moscowLocation)
}
func NextMoscowMidnight(t time.Time) time.Time { return MoscowMidnight(t).AddDate(0, 0, 1) }
func loadMoscow() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return loc
}
