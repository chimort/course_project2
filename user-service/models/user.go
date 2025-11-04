package models

type User struct {
	ID       int64    `json:"id"`        // уникальный идентификатор
	Name     string   `json:"name"`      // имя или ник
	Age      int      `json:"age"`       // возраст
	Gender   string   `json:"gender"`    // "male", "female", "other"
	Language string   `json:"language"`  // язык общения
	Tags     []string `json:"tags"`      // интересы или темы
	Online   bool     `json:"online"`    // онлайн-статус
}

