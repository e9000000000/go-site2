package main

import (
	"html/template"
	"log"
	"net/http"
)

type Middleware func(http.ResponseWriter, *http.Request, *Context)
type PathHandler func(http.ResponseWriter, *http.Request, *Context) (int, string)
type actualHandler func(w http.ResponseWriter, req *http.Request)

type Context struct {
	Identifier *Identifier
	Data       any
}

type HTTPServer struct {
	Handlers    map[string]actualHandler
	templates   *template.Template
	middlewares []Middleware
}

func NewServer() *HTTPServer {
	return &HTTPServer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
		Handlers:  make(map[string]actualHandler),
	}
}

func (serv *HTTPServer) AddMiddleware(m Middleware) {
	serv.middlewares = append(serv.middlewares, m)
}

func (serv *HTTPServer) Handle(path string, handler PathHandler) {
	actualHandler := func(w http.ResponseWriter, req *http.Request) {
		context := &Context{}
		for _, m := range serv.middlewares {
			m(w, req, context)
		}
		status, templateName := handler(w, req, context)
		serv.render(w, status, templateName, context)
	}

	serv.Handlers[path] = actualHandler
	http.HandleFunc(path, actualHandler)
}

func (serv *HTTPServer) HandleDefault(path, templateName string) {
	handler := func(w http.ResponseWriter, req *http.Request, c *Context) (int, string) {
		return 200, "base"
	}
	serv.Handle(path, handler)
}

func (serv *HTTPServer) HandleStatic(path string) {
	http.Handle(path, http.FileServer(http.Dir(".")))
}

func (serv *HTTPServer) Run(addr string) {
	log.Printf("run server on http://%s\n", addr)
	http.ListenAndServe(addr, nil)
}

func (serv *HTTPServer) render(w http.ResponseWriter, status int, templateName string, c *Context) {
	if serv.templates == nil {
		log.Println("don't forget to call InitTemplates")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t := serv.templates.Lookup(templateName)
	if t == nil {
		log.Printf("can't find tempalte by name %s\n", templateName)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	err := t.Execute(w, *c)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
