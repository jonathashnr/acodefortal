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
	ctx := app{templates, db, logger}
	// router
	router := router.NewRouter()
	router.NewRoute("GET /", ctx.homeHandler)
	router.NewRoute("GET /org/{id}", ctx.orgHandler)
	router.NewRoute("GET /cadastro", ctx.cadastroPage)
	router.NewRoute("GET /login", ctx.loginPage)
	router.NewRoute("POST /user/create", ctx.createUser)
	router.NewRoute("POST /user/login", ctx.loginUser)
	// mux and fileserver
	mux := http.NewServeMux()
	staticFilesHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	mux.Handle("/static/", staticFilesHandler)
	mux.Handle("/", router)
	// logging middleware
	loggerMiddleware := NewLoggerMiddleware(mux, logger)

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
	logger *slog.Logger
}
func (l loggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.next.ServeHTTP(w,r)
	l.logger.LogAttrs(context.Background(), slog.LevelInfo,"http request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("time_elapsed",time.Since(start).String()))
}
func NewLoggerMiddleware(nextHandler http.Handler, logger *slog.Logger) loggerMiddleware {
	if nextHandler == nil {
		return loggerMiddleware{nextHandler, slog.Default()}
	}
	return loggerMiddleware{nextHandler, logger}
}