package main

import (
	"html/template"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title string
	Text  string
}

type User struct {
	gorm.Model
	Name         string
	PasswordHash string
	IsAdmin      bool
}

var db *gorm.DB

func index(w http.ResponseWriter, req *http.Request) {
	products := make([]Post, 20)
	db.Order("created_at DESC").Limit(20).Find(&products)

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, map[string][]Post{
		"Posts": products,
	})
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Post{})

	http.HandleFunc("/", index)

	http.ListenAndServe("127.0.0.1:8000", nil)
}
