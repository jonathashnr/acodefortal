package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type app struct {
	templates *template.Template
}

func main() {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	ctx := app{templates}
	router := router{[]route{
		newRoute("GET /", ctx.homeHandler),
		newRoute("GET /org/{id}", ctx.orgHandler),
	}}
	serve := http.HandlerFunc(router.serve)
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, serve))
}