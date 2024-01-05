package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/jonathashnr/ajudafortaleza/router"
)

type app struct {
	templates *template.Template
}

func main() {
	//templates cache
	templates := template.Must(template.ParseGlob("templates/*.html"))
	ctx := app{templates}
	// router
	router := router.NewRouter()
	router.NewRoute("GET /", ctx.homeHandler)
	router.NewRoute("GET /org/{id}", ctx.orgHandler)
	// mux and fileserver
	mux := http.NewServeMux()
	staticFilesHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	mux.Handle("/static/", staticFilesHandler)
	mux.Handle("/", router)
	// logging middleware
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	loggerMiddleware := NewLoggerMiddleware(mux, logger)

	addr := ":8080"
	logger.Info("server start http://localhost"+addr)
	err := http.ListenAndServe(addr, loggerMiddleware)
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