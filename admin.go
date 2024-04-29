package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

func IdentMiddleware(w http.ResponseWriter, req *http.Request, c *Context) {
	ident := getOrCreateIdentifier(w, req)
	c.Identifier = ident
}

func AdminHandler(w http.ResponseWriter, req *http.Request, c *Context) (int, string) {
	if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			c.Data = "Невозможнос спарсить форму"
			return 400, "error"
		}

		form := req.PostForm
		if len(form["name"]) != 1 || len(form["password"]) != 1 {
			c.Data = "В форме не хватает полей"
			return 400, "error"
		}

		name := form["name"][0]
		password := form["password"][0]
		err = login(c.Identifier, name, password)
		if err != nil {
			c.Data = "Не верное имя или пароль"
			return 403, "error"
		}
	}

	if req.Method == "DELETE" {
		logout(c.Identifier)
	}

	if c.Identifier.User == nil {
		return 200, "admin-login"
	} else {
		return 200, "admin"
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
