package matching

import (
	"math"
	"strings"
)

type UserProfile struct {
	ID       string
	Age      int
	Hobbies  []string
	Language string
}

type Preferences struct {
	WeightLanguage  float64
	WeightHobbies   float64
	WeightAge       float64

	HobbyTopic      string
	HobbyTopicBoost float64

	DesiredLanguage string
	AgeTolerance    float64
}

type ScoreBreakdown struct {
	RawLanguage   float64
	RawHobbies    float64
	RawAge        float64
	WeightedLang  float64
	WeightedHobby float64
	WeightedAge   float64
	Total         float64
}

var DefaultWeights = struct{ Lang, Hobby, Age float64 }{Lang: 0.45, Hobby: 0.35, Age: 0.20}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func containsIgnoreCase(slice []string, s string) bool {
	s = strings.ToLower(s)
	for _, it := range slice {
		if strings.ToLower(it) == s {
			return true
		}
	}
	return false
}

func unionSize(a, b []string) int {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, s := range a {
		set[strings.ToLower(s)] = struct{}{}
	}
	for _, s := range b {
		set[strings.ToLower(s)] = struct{}{}
	}
	return len(set)
}

func intersectSize(a, b []string) int {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[strings.ToLower(s)] = struct{}{}
	}
	count := 0
	for _, s := range b {
		if _, ok := set[strings.ToLower(s)]; ok {
			count++
		}
	}
	return count
}

func normalizeWeights(prefs Preferences) (wLang, wHobby, wAge float64) {
	if prefs.WeightLanguage > 0 {
		wLang = prefs.WeightLanguage
	} else {
		wLang = DefaultWeights.Lang
	}
	if prefs.WeightHobbies > 0 {
		wHobby = prefs.WeightHobbies
	} else {
		wHobby = DefaultWeights.Hobby
	}
	if prefs.WeightAge > 0 {
		wAge = prefs.WeightAge
	} else {
		wAge = DefaultWeights.Age
	}
	sum := wLang + wHobby + wAge
	if sum == 0 {
		wLang = DefaultWeights.Lang
		wHobby = DefaultWeights.Hobby
		wAge = DefaultWeights.Age
		sum = wLang + wHobby + wAge
	}
	return wLang / sum, wHobby / sum, wAge / sum
}

func CompatibilityDynamic(me UserProfile, candidate UserProfile, prefs Preferences) ScoreBreakdown {
	wLang, wHobby, wAge := normalizeWeights(prefs)

	var rawLang float64
	if prefs.DesiredLanguage != "" {
		if strings.EqualFold(candidate.Language, prefs.DesiredLanguage) {
			rawLang = 1.0
		} else {
			rawLang = 0.0
		}
	} else {
		if me.Language != "" && strings.EqualFold(me.Language, candidate.Language) {
			rawLang = 1.0
		} else {
			rawLang = 0.0
		}
	}

	var rawHobby float64
	u := unionSize(me.Hobbies, candidate.Hobbies)
	if u == 0 {
		rawHobby = 0
	} else {
		rawHobby = float64(intersectSize(me.Hobbies, candidate.Hobbies)) / float64(u)
	}
	if prefs.HobbyTopic != "" {
		if containsIgnoreCase(candidate.Hobbies, prefs.HobbyTopic) {
			rawHobby += prefs.HobbyTopicBoost
			if rawHobby > 1 {
				rawHobby = 1
			}
		}
	}

	sigma := prefs.AgeTolerance
	if sigma <= 0 {
		sigma = 10.0
	}
	diff := float64(absInt(me.Age - candidate.Age))
	rawAge := math.Exp(- (diff*diff) / (2 * sigma * sigma))

	wLangC := rawLang * wLang
	wHobbyC := rawHobby * wHobby
	wAgeC := rawAge * wAge
	total := wLangC + wHobbyC + wAgeC

	return ScoreBreakdown{
		RawLanguage:   rawLang,
		RawHobbies:    rawHobby,
		RawAge:        rawAge,
		WeightedLang:  wLangC,
		WeightedHobby: wHobbyC,
		WeightedAge:   wAgeC,
		Total:         total,
	}
}

type scored struct {
	id    string
	score float64
	bd    ScoreBreakdown
}
