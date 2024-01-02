package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

type appContext struct {
	templates *template.Template
}

func main() {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	ctx := appContext{templates}
	mux := http.NewServeMux()
	mux.HandleFunc("/", ctx.homeHandler)
	newRoute("GET /users/{id}", ctx.homeHandler)
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (c *appContext)homeHandler (w http.ResponseWriter, r *http.Request) {
	c.templates.ExecuteTemplate(w, "main", nil)
}

type route struct {
	method string
	params []string
	regex *regexp.Regexp
	handler http.HandlerFunc
}

func newRoute(pattern string, handler http.HandlerFunc) route {
	validadeRegex := regexp.MustCompile(`^(GET|POST|PUT|DELETE) (\/(?:[^\/\s]\/?)*)$`)
	patternSubstrings := validadeRegex.FindStringSubmatch(pattern)
	if len(patternSubstrings) != 3 {
		panic("erro no processamento da rota: " + pattern)
	}
	method := patternSubstrings[1]
	paramRegex := regexp.MustCompile(`{([^\/\s]+)}`)
	paramSubstrings := paramRegex.FindAllStringSubmatch(patternSubstrings[2], -1)
	var params []string
	for _, paramSlice := range paramSubstrings {
		params = append(params, paramSlice[1])
	}
	// Por padrão, esse roteador considerada paths terminados com ou sem "/"
	// equivalentes, para mudar esse comportamento edite as duas linhas abaixo
	replacedPath := strings.TrimRight(paramRegex.ReplaceAllString(patternSubstrings[2],`([^\/\s]+)`), "/")
	pathRegex := regexp.MustCompile(fmt.Sprintf(`^%v/?$`, replacedPath))

	return route{method, params, pathRegex, handler}
}
