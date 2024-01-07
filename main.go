package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"text/template"

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