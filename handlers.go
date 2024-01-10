package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

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

func (a *app)superPage (w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "superuser", nil)
}

// template de error handling
type errorTmplPipe struct {
	Status int
	Title string
	Message string
}

func (a *app)errorTmplHandler(w http.ResponseWriter, status int, message string) {
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

// new user handler
func (a *app)createUser (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao dar parse em formulario", slog.Any("err", err))
		return
	}
	name, email, password := r.FormValue("name"), r.FormValue("email"), r.FormValue("password")
	if name == "" || email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if a.model.IsUserEmailTaken(email) {
		w.WriteHeader(http.StatusBadRequest)
		a.errorTmplHandler(w, http.StatusBadRequest, "Esse email já está cadastrado.")
		return
	}

	passHashedBytes, err := bcrypt.GenerateFromPassword([]byte(password),SALT_ROUNDS)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao gerar hash pelo bcrypt", slog.Any("err", err))
		return	
	}
	passHashed := string(passHashedBytes)
	
	_, err = a.model.NewUser(name, email, passHashed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao inserir novo usuário no database", slog.Any("err", err))
		return	
	}
	http.Redirect(w,r,"/",http.StatusFound)
}

func (a *app)loginUser (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao dar parse em formulario", slog.Any("err", err))
		return
	}
	email, password := r.FormValue("email"), r.FormValue("password")
	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err:= a.model.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			a.errorTmplHandler(w, http.StatusUnauthorized, "Sua autenticação falhou, algum dos campos está incorreto.")
			return	
		}
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao acessar database", slog.Any("err", err))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		a.errorTmplHandler(w, http.StatusUnauthorized, "Sua autenticação falhou, algum dos campos está incorreto.")
		return
	}
	token, err := a.model.NewSession(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("erro ao acessar database", slog.Any("err", err))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session_cookie",
		Value: token,
		Path: "/",
		HttpOnly: true,
	})
	http.Redirect(w,r,"/",http.StatusFound)
}