package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	"github.com/jonathashnr/acodefortal/database"
	"github.com/jonathashnr/acodefortal/router"
	_ "github.com/mattn/go-sqlite3"
)

type app struct {
	templates *template.Template
	model *database.Model
	logger *slog.Logger
}

func main() {
	ops := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, ops))
	//templates cache
	templates := template.Must(template.ParseGlob("templates/*.html"))
	db, err := sql.Open("sqlite3", "./database/af.db")
	if err != nil {
		logger.Error(err.Error())
	}
	defer db.Close()
	app := app{templates, database.NewModel(db), logger}
	// router
	router := router.NewRouter()
	router.NewRoute("GET /", app.homeHandler)
	router.NewRoute("GET /org/{id}", app.orgHandler)
	router.NewRoute("GET /protected", protected(app.protectedPage,0))
	router.NewRoute("GET /superuser", protected(app.superPage,5))
	router.NewRoute("GET /usuario/entrar", app.loginPage)
	router.NewRoute("POST /usuario/entrar", app.loginUser)
	router.NewRoute("GET /usuario/sair", app.logout)
	router.NewRoute("GET /usuario/cadastrar", app.cadastroPage)
	router.NewRoute("POST /usuario/cadastrar", app.createUser)
	// mux and fileserver
	mux := http.NewServeMux()
	staticFilesHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	mux.Handle("/static/", staticFilesHandler)
	mux.Handle("/", router)
	// auth middleware
	authMiddleware := app.NewAuthMiddleware(mux)
	// err and log middleware
	middleware := app.NewMiddleware(authMiddleware)

	addr := ":8080"
	logger.Info("server start http://localhost"+addr)
	err = http.ListenAndServe(addr, middleware)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}