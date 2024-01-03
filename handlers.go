package main

import (
	"fmt"
	"net/http"
)

func (a *app)homeHandler (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "main", nil)
}

func (a *app)orgHandler (w http.ResponseWriter, r *http.Request) {
	id := PathValue(r, "id")
	fmt.Fprintf(w, "O id da org é: %v", id)
}