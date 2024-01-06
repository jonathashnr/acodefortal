package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	// if message == "" {
	// 	switch status {
	// 	case http.StatusBadRequest:
	// 		message = "Requisição inválida"
	// 	case http.StatusUnauthorized:
	// 		message = "Não autenticado"
	// 	case http.StatusForbidden:
	// 		message = "Proibido/Não autorizado"
	// 	case http.StatusNotFound:
	// 		message = "Não encontrado"
	// 	}
	// }
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

func (a *app)loginUser (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.logger.Error("erro ao dar parse em form", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	email, password := r.FormValue("email"), r.FormValue("password")
	if email == "" || password == "" {
		a.errorTmplHandler(w, "Campos ausentes", http.StatusBadRequest)
		return
	}
	bcryptTimer := time.Now()
	passHashedBytes, err := bcrypt.GenerateFromPassword([]byte(password),14)
	if err != nil {
		a.logger.Error("erro ao gerar senha via bcrypt", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return	
	}
	a.logger.Debug("elapsed time", slog.String("elapsed_time", time.Since(bcryptTimer).String()))
	passHashed := string(passHashedBytes)

	var userId int
	err = a.db.QueryRow("SELECT id, FROM user WHERE email = ? AND password = ?", email, passHashed).Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			a.errorTmplHandler(w, "Autenticação falhou", http.StatusUnauthorized)
			return	
		}
		a.logger.Error("erro ao acessar do database", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	token := uuid.NewString()
	expires := time.Now().Add(300 * time.Second)
	_, err = a.db.Exec("INSERT INTO session(token,user_id,valid_until) VALUES(?,?,?)", token,userId,expires.String())
	if err != nil {
		a.logger.Error("erro ao acessar do database", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session_cookie",
		Value: token,
		Expires: expires,
		HttpOnly: true,
	})
	a.templates.ExecuteTemplate(w, "main", nil)
}