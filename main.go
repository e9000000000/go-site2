package main

import (
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func BaseHandler(w http.ResponseWriter, req *http.Request, c *Context) (int, string) {
	return 200, "base"
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

	serv := NewServer()

	serv.HandleStatic("/staitc/")

	serv.AddMiddleware(IdentMiddleware)

	serv.Handle("/", BaseHandler)
	serv.Handle("/posts", PostsHandler)
	serv.Handle("/admin", AdminHandler)

	serv.Run("127.0.0.1:8000")
}
