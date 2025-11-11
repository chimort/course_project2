package converter

import (
	"github.com/chimort/course_project2/api/proto/sharedpb"
	"github.com/chimort/course_project2/iternal/user/models"
)

// ------------------ Languages ------------------
func ToPbLanguages(langs []models.UserLanguage) []*sharedpb.Language {
	out := make([]*sharedpb.Language, len(langs))
	for i, l := range langs {
		out[i] = &sharedpb.Language{
			Name:  string(l.Language),
			Level: languageLevelToProto(l.Level),
		}
	}
	return out
}

func FromPbLanguages(langs []*sharedpb.Language) []models.UserLanguage {
	out := make([]models.UserLanguage, len(langs))
	for i, l := range langs {
		out[i] = models.UserLanguage{
			Language: models.Language(l.Name),
			Level:    languageLevelFromProto(l.Level),
		}
	}
	return out
}

func languageLevelToProto(level models.LanguageLevel) sharedpb.LanguageLevel {
	switch level {
	case "NATIVE":
		return sharedpb.LanguageLevel_NATIVE
	case "MEDIUM":
		return sharedpb.LanguageLevel_MEDIUM
	case "LOW":
		return sharedpb.LanguageLevel_LOW
	default:
		return sharedpb.LanguageLevel_LOW
	}
}

func languageLevelFromProto(level sharedpb.LanguageLevel) models.LanguageLevel {
	switch level {
	case sharedpb.LanguageLevel_NATIVE:
		return "NATIVE"
	case sharedpb.LanguageLevel_MEDIUM:
		return "MEDIUM"
	case sharedpb.LanguageLevel_LOW:
		return "LOW"
	default:
		return "LOW"
	}
}

// ------------------ Interests ------------------
func ToPbInterests(ints []models.UserInterest) []*sharedpb.Interests {
	out := make([]*sharedpb.Interests, len(ints))
	for i, v := range ints {
		out[i] = &sharedpb.Interests{Name: string(v.Interest)}
	}
	return out
}

func FromPbInterests(ints []*sharedpb.Interests) []models.UserInterest {
	out := make([]models.UserInterest, len(ints))
	for i, v := range ints {
		out[i] = models.UserInterest{Interest: models.Interests(v.Name)}
	}
	return out
}
