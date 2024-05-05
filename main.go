package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB(dbFile, adminName, adminPassword string) {
	var err error
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Post{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Identifier{})

	CreateAdmin(adminName, adminPassword)
}

func initServer() *HTTPServer {
	serv := NewServer()

	serv.HandleStatic("/static/")

	serv.AddMiddleware(IdentMiddleware)

	serv.HandleDefault("/", "base")
	serv.Handle("/posts", PostsHandler)
	serv.Handle("/admin", AdminHandler)
	serv.Handle("/admin-posts", AdminPostsHandler)

	return serv
}

func main() {
	initDB("db.sqlite3", "admin", "123")
	serv := initServer()
	serv.Run("127.0.0.1:8000")
}
