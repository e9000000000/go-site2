package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type User struct {
	ID           uint   `gorm:"primarykey"`
	Name         string `gorm:"uniqueIndex"`
	PasswordHash string
}

type Identifier struct {
	ID     uint   `gorm:"primarykey"`
	UUID   string `gorm:"uniqueIndex"`
	UserID int
	User   *User `gorm:"foreignKey:UserID;references:ID"`
}

func AdminHandler(w http.ResponseWriter, req *http.Request) {
	ident := getOrCreateIdentifier(w, req)

	if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			http.Error(w, "can't parse form", http.StatusBadRequest)
			return
		}

		form := req.PostForm
		if len(form["name"]) != 1 || len(form["password"]) != 1 {
			http.Error(w, "fields are missing", http.StatusBadRequest)
			return
		}
		name := form["name"][0]
		password := form["password"][0]
		err = login(ident, name, password)
		if err != nil {
			http.Error(w, "Не верное имя или пароль", http.StatusBadRequest)
			return
		}
	}

	if req.Method == "DELETE" {
		logout(ident)
	}

	if ident == nil || ident.User == nil {
		renderTemplate(w, "admin-login.html", nil)
	} else {
		renderTemplate(w, "admin.html", nil)
	}
}

func CreateAdmin(name, password string) (*User, error) {
	user := User{
		Name:         name,
		PasswordHash: hashPassword(password),
	}
	err := db.Create(&user).Error
	if err == nil {
		return nil, fmt.Errorf("user with name %s already exists", name)
	}
	return &user, nil
}

func renderTemplate(w http.ResponseWriter, templateName string, data any) {
	t, err := template.ParseFiles("templates/base.html", "templates/"+templateName)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = t.Execute(w, data)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func hashPassword(password string) string {
	rawHash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(rawHash[:])
}

func getOrCreateIdentifier(w http.ResponseWriter, req *http.Request) *Identifier {
	var ident_uuid string

	cookies := req.Cookies()
	for _, c := range cookies {
		if c.Name != "UUID" {
			continue
		}

		ident_uuid = c.Value
		break
	}

	if ident_uuid == "" {
		ident_uuid = uuid.New().String()
		newCookie := http.Cookie{Name: "UUID", Value: ident_uuid, HttpOnly: true}
		http.SetCookie(w, &newCookie)
	}

	ident := Identifier{
		UUID: ident_uuid,
	}
	if err := db.Joins("User").Where("uuid = ?", ident_uuid).First(&ident).Error; err == nil {
		return &ident
	}

	if err := db.Create(&ident).Error; err != nil {
		panic(err)
	}

	return &ident
}

func login(ident *Identifier, name, password string) error {
	passwordHash := hashPassword(password)
	var user User // TODO should be initialized?!
	err := db.
		Where("name = ?", name).
		Where("password_hash = ?", passwordHash).
		First(&user).
		Error
	if err != nil {
		return fmt.Errorf("no user with this password and name in db")
	}

	ident.User = &user
	if err := db.Save(&ident).Error; err != nil {
		panic(err)
	}
	return nil
}

func logout(ident *Identifier) error {
	ident.User = nil
	return db.Save(&ident).Error
}
