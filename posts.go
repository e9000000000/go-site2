package main

import (
	"log"
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

func AdminPostsHandler(w http.ResponseWriter, req *http.Request, c *Context) (int, string) {
	if c.Identifier.User == nil {
		c.Data = "Только для админов"
		return 403, "error"
	}

	if req.Method == "POST" {
		newPost := Post{
			Title: req.PostFormValue("title"),
			Text:  req.PostFormValue("text"),
		}

		if err := db.Create(&newPost).Error; err != nil {
			log.Println(err)
			c.Data = "Не создается"
			return 500, "error"
		}

		return 200, "admin"
	}

	return 200, "admin-posts"
}
