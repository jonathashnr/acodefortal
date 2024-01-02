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
	router := router{[]route{
		newRoute("GET /", ctx.homeHandler),
		newRoute("GET /org/{id}", ctx.orgHandler),
	}}
	serve := http.HandlerFunc(router.serve)
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, serve))
}

func (c *appContext)homeHandler (w http.ResponseWriter, r *http.Request) {
	c.templates.ExecuteTemplate(w, "main", nil)
}

func (c *appContext)orgHandler (w http.ResponseWriter, r *http.Request) {
	id := PathValue(r, "id")
	fmt.Fprintf(w, "O id da org é: %v", id)
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

type router struct {
	routes []route
}
type ctxkey struct {}

func (router *router) serve(w http.ResponseWriter, r *http.Request) {
	var allowedMethods []string
	for _, route:= range router.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allowedMethods = append(allowedMethods, route.method)
				continue
			}
			paramValues := matches[1:]
			// Não é para disparar nunca
			if(len(route.params) != len(paramValues)) {
				log.Printf("discrepância inesperada entre tamanho dos slices de params. route: %v. matches:%v", route.params, matches[1:])
				continue
			}
			param := make(map[string]string)
			for i, key := range route.params {
				param[key] = paramValues[i]
			}
			context := context.WithValue(r.Context(), ctxkey{}, param)
			route.handler(w,r.WithContext(context))
			return
		}
	}
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		return
	}
	http.Error(w, "Página não encontrada", http.StatusNotImplemented)
}

func PathValue(r *http.Request, key string) string {
	param := r.Context().Value(ctxkey{}).(map[string]string)
	return param[key]
}