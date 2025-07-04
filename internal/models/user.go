package models

type User struct {
    Nickname string `json:"nickname"`
    Fullname string `json:"fullname"`
    About    string `json:"about"`
    Email    string `json:"email"`
}

type NewUser struct {
    Fullname string `json:"fullname"`
    About    string `json:"about"`
    Email    string `json:"email"`
}

type Users []*User

type UserUpdate struct {
    Fullname string `json:"fullname"`
    About    string `json:"about"`
    Email    string `json:"email"`
}