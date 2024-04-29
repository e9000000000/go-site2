package main

import (
	"net/http"
)

type Post struct {
	ID    uint `gorm:"primarykey"`
	Title string
	Text  string
}

func PostsHandler(w http.ResponseWriter, req *http.Request, c *Context) (int, string) {
	posts := make([]Post, 20)
	db.Order("created_at DESC").Limit(20).Find(&posts)

	c.Data = posts
	return 200, "posts"
}
