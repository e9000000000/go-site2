package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

var serv *HTTPServer

func TestMain(m *testing.M) {
	initDB("testdb.sqlite3", "admimimin", "1215")
	serv = initServer()

	status := m.Run()

	os.Remove("testdb.sqlite3")
	os.Exit(status)
}

func adminGetAuthCookieJar(u string) *cookiejar.Jar {
	handler := serv.Handlers["/admin"]
	server := httptest.NewServer(http.HandlerFunc(handler))
	resp, _ := http.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewReader([]byte("name=admimimin&password=1215")))

	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name != "UUID" {
			continue
		}
		authCookie = c
		break
	}
	if authCookie == nil {
		panic("no auth cookie")
	}

	cookieJar, _ := cookiejar.New(nil)
	parsedUrl, _ := url.Parse(u)
	cookieJar.SetCookies(parsedUrl, []*http.Cookie{authCookie})
	return cookieJar
}

func TestBase(t *testing.T) {
	handler := serv.Handlers["/"]
	server := httptest.NewServer(http.HandlerFunc(handler))

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status code is %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
	}

	respText := string(b)
	if !strings.Contains(strings.ToLower(respText), "добро пожаловать") {
		t.Errorf("no good words on main page")
	}
}

func TestPosts(t *testing.T) {
	testPost := Post{
		Title: "test post 1",
		Text:  "test ppost 1 text",
	}
	if err := db.Create(&testPost).Error; err != nil {
		t.Error(err)
	}

	handler := serv.Handlers["/posts"]
	server := httptest.NewServer(http.HandlerFunc(handler))

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Error("status code is not 200")
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
	}

	respText := string(b)
	if !strings.Contains(respText, testPost.Title) {
		t.Error("no post title in response text")
	}
	if !strings.Contains(respText, testPost.Text) {
		t.Error("no post text in response text")
	}
}

func TestAdmin(t *testing.T) {
	handler := serv.Handlers["/admin"]
	server := httptest.NewServer(http.HandlerFunc(handler))

	cookieJar, _ := cookiejar.New(nil)
	c := http.Client{
		Jar: cookieJar,
	}

	// GET
	resp, err := c.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Error("status code is not 200")
	}

	b, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	respText := string(b)

	if !strings.Contains(strings.ToLower(respText), "войдите") {
		t.Error("it should show login form")
	}

	// POST trying to log in with wrong data
	data := []byte("name=admimimin&password=2313")
	resp, err = c.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewReader(data))

	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode/100 != 4 {
		t.Error("status code is not 4xx")
	}

	// POST trying to log in with right data
	data = []byte("name=admimimin&password=1215")
	resp, err = c.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewReader(data))

	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode/100 != 2 {
		t.Error("status code is not 2xx")
		return
	}

	// GET check we are logined in
	resp, err = c.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Error("status code is not 200")
	}

	b, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()
	respText = string(b)

	if !strings.Contains(strings.ToLower(respText), "добро пожаловать в админку") {
		t.Error("it should show an actual admin")
	}

	// DELETE logout
	req, _ := http.NewRequest("DELETE", server.URL, nil)
	resp, err = c.Do(req)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Error("status code is not 200")
	}

	// GET should be login form again
	resp, err = c.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Error("status code is not 200")
	}

	b, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()
	respText = string(b)

	if !strings.Contains(strings.ToLower(respText), "войдите") {
		t.Error("it should show login form")
	}
}

func TestPostsAdmin(t *testing.T) {
	// TODO complete this test
	handler := serv.Handlers["/admin-posts"]
	server := httptest.NewServer(http.HandlerFunc(handler))

	data := []byte("title=fweiojfoweij&text=ofwejiofiewjowefjifweo")

	resp, err := http.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode / 100 != 4 {
		t.Error("should be unawailable for not admins")
	}

	c := http.Client{
		Jar: adminGetAuthCookieJar(server.URL),
	}

	resp, err = c.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode / 100 != 2 {
		t.Error("not 2xx status code")
	}

	post := Post{}
	err = db.Order("id DESC").Limit(1).First(&post).Error
	if err != nil {
		t.Error(err)
	}

	if post.Title != "fweiojfoweij" || post.Text != "ofwejiofiewjowefjifweo" {
		t.Error("post not saved, or has wrong data")
	}
}
