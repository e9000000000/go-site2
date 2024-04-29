package main

import (
	"html/template"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Post struct {
	ID    uint `gorm:"primarykey"`
	Title string
	Text  string
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	products := make([]Post, 20)
	db.Order("created_at DESC").Limit(20).Find(&products)

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
		return
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

	CreateAdmin("admin", "123")

	db.AutoMigrate(&Post{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Identifier{})

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/admin", AdminHandler)

	http.ListenAndServe("127.0.0.1:8000", nil)
}
