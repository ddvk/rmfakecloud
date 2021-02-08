package ui

import "time"

type loginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type DocumentTree struct {
	Entries []Entry
}

type Entry interface {
}
type Directory struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Documents []Document `json:"documents"`
}

// Document is a single document
type Document struct {
	ID       string `json:id`
	Name     string `json:name`
	ImageUrl string `json:imageUrl`
	ParentId string `json:parentId`
}

// DocumentList is a list of documents
type DocumentList struct {
	Documents []Document `json:documents`
}
type user struct {
	ID        string `json:"userid"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt time.Time
}
