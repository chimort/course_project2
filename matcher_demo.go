package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// User описывает профиль
type User struct {
	ID       string
	Age      int
	Hobbies  []string
	Language string
	Location string
}

// Preferences — как пользователь хочет искать
type Preferences struct {
	// Если вес = 0 — будет использовано значение по умолчанию из DefaultWeights.
	WeightLanguage float64 // желаемый вес языка
	WeightHobbies  float64 // желаемый вес хобби
	WeightAge      float64 // желаемый вес возраста
	WeightLocation float64 // желаемый вес локации

	// Специальные опции
	HobbyTopic     string  // если задано, кандидаты, у которых есть такой хобби-топик, получают буст
	HobbyTopicBoost float64 // на сколько (0..1) увеличить hobby score при совпадении топика (будет лимитироваться до 1.0)

	DesiredLanguage string  // если задано, язык считается относительно этой цели (например, вы хотите найти носителя Spanish)
	AgeTolerance    float64 // sigma для gaussian age score (чем меньше — тем строже). По умолчанию 10.
}

// ScoreBreakdown — компонентные и итоговые взвешенные значения
type ScoreBreakdown struct {
	RawLanguage   float64 // сырые (0..1)
	RawHobbies    float64
	RawAge        float64
	RawLocation   float64
	WeightedLang  float64 // contribution = Raw * weight
	WeightedHobby float64
	WeightedAge   float64
	WeightedLoc   float64
	Total         float64
}

// DefaultWeights — если пользователь не задаёт веса, эти используются
var DefaultWeights = struct {
	Lang, Hobby, Age, Loc float64
}{Lang: 0.35, Hobby: 0.35, Age: 0.2, Loc: 0.1}

// --- вспомогательные функции ---
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

// normalizeWeights: собирает веса из prefs (если заданы) или из defaults, возвращает нормализованные веса, сумма = 1.
func normalizeWeights(prefs Preferences) (wLang, wHobby, wAge, wLoc float64) {
	// возьмём либо пользовательское значение, либо дефолт
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
	if prefs.WeightLocation > 0 {
		wLoc = prefs.WeightLocation
	} else {
		wLoc = DefaultWeights.Loc
	}
	sum := wLang + wHobby + wAge + wLoc
	// на всякий случай — если сумма ноль (краевой случай), восстановим дефолты
	if sum == 0 {
		wLang = DefaultWeights.Lang
		wHobby = DefaultWeights.Hobby
		wAge = DefaultWeights.Age
		wLoc = DefaultWeights.Loc
		sum = wLang + wHobby + wAge + wLoc
	}
	// нормируем
	return wLang / sum, wHobby / sum, wAge / sum, wLoc / sum
}

// CompatibilityDynamic считает скор и возвращает breakdown.
// Логика:
// - language: если prefs.DesiredLanguage задан — проверяем совпадение кандидата с ним (удобно для "хочу учить Spanish")
//   иначе считаем совпадение по общему языку (user.Language == candidate.Language).
// - hobbies: базовый Jaccard (intersect/union). Если prefs.HobbyTopic задан и кандидат имеет этот топик,
//   то hobbyScore увеличивается на HobbyTopicBoost (ограничено 1.0).
// - age: gaussian decay: s = exp(- (diff^2) / (2*sigma^2)), где sigma = prefs.AgeTolerance (по умолчанию 10).
// - location: 1 если совпадает, иначе 0.
// Затем итог = sum_i weight_i * raw_i, где веса нормализованы (сумма=1).
func CompatibilityDynamic(user User, candidate User, prefs Preferences) ScoreBreakdown {
	// Нормализовать веса
	wLang, wHobby, wAge, wLoc := normalizeWeights(prefs)

	// LANGUAGE raw score
	var rawLang float64
	if prefs.DesiredLanguage != "" {
		// цель — найти носителя DesiredLanguage
		if strings.EqualFold(candidate.Language, prefs.DesiredLanguage) {
			rawLang = 1.0
		} else {
			rawLang = 0.0
		}
	} else {
		// сравнение по общему языку (если у пользователя он известен)
		if user.Language != "" && strings.EqualFold(user.Language, candidate.Language) {
			rawLang = 1.0
		} else {
			rawLang = 0.0
		}
	}

	// HOBBIES raw score: Jaccard similarity
	var rawHobby float64
	u := unionSize(user.Hobbies, candidate.Hobbies)
	if u == 0 {
		rawHobby = 0
	} else {
		rawHobby = float64(intersectSize(user.Hobbies, candidate.Hobbies)) / float64(u)
	}
	// приоритетный топик в хобби — добавить буст
	if prefs.HobbyTopic != "" {
		// если кандидат имеет этот топик (чувствительно к регистру через containsIgnoreCase)
		if containsIgnoreCase(candidate.Hobbies, prefs.HobbyTopic) {
			rawHobby = rawHobby + prefs.HobbyTopicBoost
			if rawHobby > 1 {
				rawHobby = 1
			}
		}
	}

	// AGE raw score: gaussian / exp decay
	sigma := prefs.AgeTolerance
	if sigma <= 0 {
		sigma = 10.0 // дефолт
	}
	diff := float64(absInt(user.Age-candidate.Age))
	rawAge := math.Exp(- (diff*diff) / (2 * sigma * sigma))

	// LOCATION raw score
	var rawLoc float64
	if user.Location != "" && strings.EqualFold(user.Location, candidate.Location) {
		rawLoc = 1.0
	} else {
		rawLoc = 0.0
	}

	// weighted contributions
	wLangC := rawLang * wLang
	wHobbyC := rawHobby * wHobby
	wAgeC := rawAge * wAge
	wLocC := rawLoc * wLoc
	total := wLangC + wHobbyC + wAgeC + wLocC

	return ScoreBreakdown{
		RawLanguage:   rawLang,
		RawHobbies:    rawHobby,
		RawAge:        rawAge,
		RawLocation:   rawLoc,
		WeightedLang:  wLangC,
		WeightedHobby: wHobbyC,
		WeightedAge:   wAgeC,
		WeightedLoc:   wLocC,
		Total:         total,
	}
}

func main() {
	// основной пользователь
	me := User{
		ID:       "me",
		Age:      28,
		Hobbies:  []string{"movies", "reading", "cooking"},
		Language: "English",
		Location: "NY",
	}

	// кандидаты
	others := []User{
		{ID: "alice", Age: 27, Hobbies: []string{"movies", "reading"}, Language: "English", Location: "NY"},
		{ID: "bob", Age: 35, Hobbies: []string{"cooking", "gaming"}, Language: "Spanish", Location: "LA"},
		{ID: "carol", Age: 22, Hobbies: []string{"reading", "yoga", "cooking"}, Language: "English", Location: "NY"},
		{ID: "dave", Age: 31, Hobbies: []string{"movies", "golf"}, Language: "English", Location: "NY"},
		{ID: "eve", Age: 29, Hobbies: []string{"movies", "cooking", "reading", "gaming"}, Language: "Spanish", Location: "SF"},
	}

	// Пример 1: базовые веса (дефолт)
	fmt.Println("=== Example 1: Default weights (no special preferences) ===")
	defaultPrefs := Preferences{} // пустой — будут дефолтные веса
	printMatches(me, others, defaultPrefs)

	// Пример 2: пользователь хочет именно поговорить о фильмах — усиливаем хобби + задаём HobbyTopic "movies"
	fmt.Println("\n=== Example 2: HobbyTopic = \"movies\" with boost, hobby made more important ===")
	prefsMovies := Preferences{
		HobbyTopic:      "movies",
		HobbyTopicBoost: 0.25, // если у кандидата есть movies — увеличим hobby raw на 0.25 (макс 1.0)
		// поднимем вес хобби явно
		WeightHobbies: 0.6,
		WeightLanguage: 0.15,
		WeightAge: 0.2,
		WeightLocation: 0.05,
		AgeTolerance: 8, // чуть строже по возрасту
	}
	printMatches(me, others, prefsMovies)

	// Пример 3: пользователь хочет выучить Spanish — задаём DesiredLanguage
	fmt.Println("\n=== Example 3: DesiredLanguage = \"Spanish\" (ищем носителей Spanish) ===")
	prefsLearnSpanish := Preferences{
		DesiredLanguage: "Spanish",
		// язык очень важен для этой цели:
		WeightLanguage: 0.6,
		WeightHobbies:  0.2,
		WeightAge:      0.15,
		WeightLocation: 0.05,
		AgeTolerance:   12,
	}
	printMatches(me, others, prefsLearnSpanish)
}

// helper: печать матчей (сортировка по total desc)
func printMatches(me User, others []User, prefs Preferences) {
	type r struct {
		ID   string
		Bd   ScoreBreakdown
	}
	var res []r
	for _, cand := range others {
		b := CompatibilityDynamic(me, cand, prefs)
		res = append(res, r{ID: cand.ID, Bd: b})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Bd.Total > res[j].Bd.Total
	})

	// покажем используемые веса
	wLang, wHobby, wAge, wLoc := normalizeWeights(prefs)
	fmt.Printf("Normalized weights: language=%.2f, hobbies=%.2f, age=%.2f, location=%.2f\n", wLang, wHobby, wAge, wLoc)
	if prefs.HobbyTopic != "" {
		fmt.Printf("HobbyTopic='%s', HobbyTopicBoost=%.2f\n", prefs.HobbyTopic, prefs.HobbyTopicBoost)
	}
	if prefs.DesiredLanguage != "" {
		fmt.Printf("DesiredLanguage='%s'\n", prefs.DesiredLanguage)
	}
	fmt.Println("Matches (sorted):")
	for _, r := range res {
		b := r.Bd
		fmt.Printf("- %s: total=%.3f  (raw: lang=%.2f hobby=%.2f age=%.2f loc=%.2f)  (weighted: lang=%.3f hobby=%.3f age=%.3f loc=%.3f)\n",
			r.ID,
			b.Total,
			b.RawLanguage, b.RawHobbies, b.RawAge, b.RawLocation,
			b.WeightedLang, b.WeightedHobby, b.WeightedAge, b.WeightedLoc,
		)
	}
}
