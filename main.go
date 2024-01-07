package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/jonathashnr/ajudafortaleza/router"
	_ "github.com/mattn/go-sqlite3"
)

type app struct {
	templates *template.Template
	db *sql.DB
	logger *slog.Logger
}

func main() {
	ops := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, ops))
	//templates cache
	templates := template.Must(template.ParseGlob("templates/*.html"))
	db, err := sql.Open("sqlite3", "./bin/mydb.db")
	if err != nil {
		logger.Error(err.Error())
	}
	defer db.Close()
	app := app{templates, db, logger}
	// router
	router := router.NewRouter()
	router.NewRoute("GET /", app.homeHandler)
	router.NewRoute("GET /org/{id}", app.orgHandler)
	router.NewRoute("GET /cadastro", app.cadastroPage)
	router.NewRoute("GET /login", app.loginPage)
	router.NewRoute("GET /protected", app.protected(app.protectedPage))
	router.NewRoute("POST /user/create", app.createUser)
	router.NewRoute("POST /user/login", app.loginUser)
	// mux and fileserver
	mux := http.NewServeMux()
	staticFilesHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	mux.Handle("/static/", staticFilesHandler)
	mux.Handle("/", router)
	// auth middleware
	authMiddleware := app.NewAuthMiddleware(mux)
	// logging middleware
	loggerMiddleware := app.NewLoggerMiddleware(authMiddleware)

	addr := ":8080"
	logger.Info("server start http://localhost"+addr)
	err = http.ListenAndServe(addr, loggerMiddleware)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

type loggerMiddleware struct{
	next http.Handler
	*app
}
func (l loggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.next.ServeHTTP(w,r)
	l.logger.LogAttrs(context.Background(), slog.LevelInfo,"http request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("time_elapsed",time.Since(start).String()))
}
func (a *app) NewLoggerMiddleware(nextHandler http.Handler) loggerMiddleware {
	return loggerMiddleware{nextHandler, a}
}

type authMiddleware struct{
	next http.Handler
	*app
}
type authKey struct {}
type sessionInfo struct {
	auth bool
	userId int
}
func (auth authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_cookie")
	if err != nil {
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	token := sessionCookie.Value
	userId, err := auth.getUserIdFromActiveSession(token)
	if err != nil {
		if err != sql.ErrNoRows {
			auth.logger.Error("erro ao acessar database", slog.String("errMsg",err.Error()))
		}
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	ctx := context.WithValue(r.Context(), authKey{},sessionInfo{true,userId})
	auth.next.ServeHTTP(w,r.WithContext(ctx))
	// essa função escreve no db em TODA requisição de users
	// autenticados e na minha maquina adiciona 8-10ms a toda req,
	// será que devia fazer um cache?
	auth.prolongSession(token)
}

func (a *app)NewAuthMiddleware(nextHandler http.Handler) authMiddleware {
	return authMiddleware{nextHandler,a}
}

func (a *app)protected(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authInfo := r.Context().Value(authKey{}).(sessionInfo)
		if authInfo.auth {
			next(w,r)
			return
		}
		a.errorTmplHandler(w,"Não autorizado", http.StatusUnauthorized)
	}
}