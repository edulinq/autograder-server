package model

type User struct {
    Email string `json:"email"`
    DisplayName string `json:"display-name"`
    Role UserRole `json:"role"`
    Pass string `json:"pass"`
}
