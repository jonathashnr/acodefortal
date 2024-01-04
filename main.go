package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/jonathashnr/ajudafortaleza/router"
)

type app struct {
	templates *template.Template
}

func main() {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	ctx := app{templates}
	mux := router.NewRouter()
	mux.NewRoute("GET /", ctx.homeHandler)
	mux.NewRoute("GET /org/{id}", ctx.orgHandler)
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, mux))
}