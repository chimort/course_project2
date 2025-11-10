package models

type UserLanguage struct {
    Language Language      `db:"language"`       
    Level    LanguageLevel `db:"proficiency_level"`
}

type UserInterest struct {
    Interest Interests `db:"interest"` 
}

type User struct {
    Username  string          `db:"username"`
    Email     string          `db:"email"`
    Password  string          `db:"password_hash"`
    Age       int             `db:"age"`
    Gender    string          `db:"gender"`
    Languages []UserLanguage  `db:"-"` 
    Interests []UserInterest  `db:"-"` 
}