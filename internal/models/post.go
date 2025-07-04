package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
    ID       int       `json:"id"`
    Parent   int       `json:"parent"`
    Author   string    `json:"author"`
    Message  string    `json:"message"`
    IsEdited bool      `json:"isEdited"`
    Forum    string    `json:"forum"`
    Thread   int       `json:"thread"`
    Created  time.Time `json:"created"`
    Path     pq.Int64Array     `json:"-"` 
}

type NewPost struct {
    Parent  int    `json:"parent"`
    Author  string `json:"author"`
    Message string `json:"message"`
}

type Posts []*Post
type NewPosts []*NewPost

type PostUpdate struct {
    Message string `json:"message"`
}

type PostFull struct {
	Post   *Post   `json:"post"`
	Author *User   `json:"author,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
}