package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type appContext struct {
	templates *template.Template
}

func main() {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "main", nil)
	})
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (c *appContext)homeHandler (w http.ResponseWriter, r *http.Request) {
	c.templates.ExecuteTemplate(w, "main", nil)
}