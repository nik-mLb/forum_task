package models

import "time"

type Thread struct {
    ID      int       `json:"id"`
    Title   string    `json:"title"`
    Author  string    `json:"author"`
    Forum   string    `json:"forum"`
    Message string    `json:"message"`
    Votes   int       `json:"votes"`
    Slug    string    `json:"slug"`
    Created time.Time `json:"created"`
}

type NewThread struct {
    Title   string `json:"title"`
    Author  string `json:"author"`
    Message string `json:"message"`
    Slug    string `json:"slug"`
    Created string `json:"created"`
}

type Threads []*Thread

type ThreadUpdate struct {
    Title   string `json:"title"`
    Message string `json:"message"`
}