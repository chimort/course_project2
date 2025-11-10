package models

type Interests string

const (
	InterestMusic  Interests = "music"
	InterestMovies Interests = "movies"
	InterestSport  Interests = "sport"
	InterestBooks  Interests = "books"
)

type Language string

const (
	LanguageRU Language = "ru"
	LanguageEN Language = "en"
)

type LanguageLevel string

const (
    LevelNative LanguageLevel = "native"
    LevelMedium LanguageLevel = "medium"
    LevelLow    LanguageLevel = "low"
)