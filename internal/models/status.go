package models

type Status struct {
	Post int `json:"post"`
	Thread int `json:"thread"`
	User	int `json:"user"`
	Forum int `json:"forum"`
}