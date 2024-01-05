package main

import (
	"fmt"
	"net/http"

	"github.com/jonathashnr/ajudafortaleza/router"
	"golang.org/x/crypto/bcrypt"
)

func (a *app)homeHandler (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "main", nil)
}

func (a *app)orgHandler (w http.ResponseWriter, r *http.Request) {
	id := router.PathValue(r, "id")
	fmt.Fprintf(w, "O id da org é: %v", id)
}
func (a *app)cadastroPage (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "cadastro", nil)
}
func (a *app)loginPage (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "login", nil)
}
type errorTmplPipe struct {
	Status int
	Message string
}
func (a *app)errorTmplHandler(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	a.templates.ExecuteTemplate(w,"error",errorTmplPipe{status,message})
}
func (a *app)createUser (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	name, email, password := r.FormValue("name"), r.FormValue("email"), r.FormValue("password")
	if name == "" || email == "" || password == "" {
		a.errorTmplHandler(w, "Campos ausentes", http.StatusBadRequest)
		return
	}

	var isEmailRegistered bool
	if err = a.db.QueryRow("SELECT COUNT(1) FROM user WHERE email = ?", email).Scan(&isEmailRegistered); err != nil {
		a.errorTmplHandler(w, "Erro ao acessar o database", http.StatusInternalServerError)
		return
	}
	if isEmailRegistered {
		a.errorTmplHandler(w, "Email já cadastrado", http.StatusBadRequest)
		return
	}

	passHashedBytes, err := bcrypt.GenerateFromPassword([]byte(password),14)
	if err != nil {
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return	
	}
	passHashed := string(passHashedBytes)
	
	_, err = a.db.Exec("INSERT INTO user(name,email,password) VALUES(?,?,?)", name,email,passHashed)
	if err != nil {
		a.errorTmplHandler(w, "Problema ao dar insert no db", http.StatusInternalServerError)
		return	
	}
	a.templates.ExecuteTemplate(w, "login", nil)
}