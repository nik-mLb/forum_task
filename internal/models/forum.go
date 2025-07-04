package models

type Forum struct {
    Title   string `json:"title"`
    User    string `json:"user"`
    Slug    string `json:"slug"`
    Posts   int    `json:"posts"`
    Threads int    `json:"threads"`
}

type NewForum struct {
    Title string `json:"title"`
    User  string `json:"user"`
    Slug  string `json:"slug"`
}