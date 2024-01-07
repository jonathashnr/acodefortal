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
const SALT_ROUNDS int = 12

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
func (a *app)protectedPage (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "protected", nil)
}
type errorTmplPipe struct {
	Status int
	Title string
	Message string
}
func (a *app)errorTmplHandler(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	var title, defaultMsg string
	switch status {
	case http.StatusBadRequest:
		title = "Requisição Inválida"
		defaultMsg = "Parece que há um problema com o que você nos enviou. Talvez algum campo faltando ou formatação incorreta."
	case http.StatusUnauthorized:
		title = "Não Autenticado"
		defaultMsg = "Você precisa fazer login antes de acessar essa página/recurso."
	case http.StatusForbidden:
		title = "Proibido/Não Autorizado"
		defaultMsg = "Você não tem autorização para acessar essa página/recurso."
	case http.StatusNotFound:
		title = "Não Encontrado"
		defaultMsg = "Vish, não tem nada aqui. :("
	case http.StatusMethodNotAllowed:
		title = "Método Não Permitido"
		defaultMsg = "Que diabos você tá fazendo?"
	case http.StatusInternalServerError:
		title = "Erro Interno no Servidor"
		defaultMsg = "Algo inesperado aconteceu. Boa sorte pra mim."
	default:
		title = "Erro"
		defaultMsg = "Nunca nem vi esse erro na vida."
	}
	if message == "" {
		message = defaultMsg
	}
	a.templates.ExecuteTemplate(w,"error",errorTmplPipe{status,title,message})
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

	passHashedBytes, err := bcrypt.GenerateFromPassword([]byte(password),SALT_ROUNDS)
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

	var userId int
	var hashedPass string
	err = a.db.QueryRow("SELECT id, password FROM user WHERE email = ?", email).Scan(&userId,&hashedPass)
	if err != nil {
		if err == sql.ErrNoRows {
			a.errorTmplHandler(w, "Autenticação falhou", http.StatusUnauthorized)
			return	
		}
		a.logger.Error("erro ao acessar do database", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPass),[]byte(password))
	if err != nil {
		a.errorTmplHandler(w, "Autenticação falhou", http.StatusUnauthorized)
		return
	}
	token, err := a.newUserSession(userId)
	if err != nil {
		a.logger.Error("erro ao acessar do database", slog.String("errMsg", err.Error()))
		a.errorTmplHandler(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session_cookie",
		Value: token,
		Path: "/",
		HttpOnly: true,
	})
	a.templates.ExecuteTemplate(w, "main", nil)
}

// models?
func (a *app) newUserSession(userId int) (token string, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	token = uuid.String()
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err = a.db.Exec("INSERT INTO session(token,user_id,expires) VALUES(?,?,?)", token, userId, expires)
	if err != nil {
		return "", err
	}
	return token, err
}

// func (a *app) isSessionActive(token string) bool {
// 	var isIt bool
// 	_ = a.db.QueryRow("SELECT COUNT(1) FROM session WHERE token = ? AND expires >= ?", token,time.Now().Unix()).Scan(&isIt)
// 	return isIt
// }
func (a *app) prolongSession(token string) error {
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err := a.db.Exec("UPDATE session SET expires = ? WHERE token = ?", expires, token)
	return err
}
func (a *app) getUserIdFromActiveSession(token string) (int, error) {
	var userId int
	err := a.db.QueryRow("SELECT user_id FROM session WHERE token = ? AND expires >= ?", token,time.Now().Unix()).Scan(&userId)
	return userId, err
}