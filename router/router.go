package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type route struct {
	method string
	params []string
	regex *regexp.Regexp
	handler http.HandlerFunc
}

type router struct {
	routes []route
}

type ParamKey struct {}

func NewRouter() router {
	return router{}
}

func (router *router)NewRoute(pattern string, handler http.HandlerFunc) {
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

	router.routes = append(router.routes, route{method, params, pathRegex, handler})
}

func (router router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			context := context.WithValue(r.Context(), ParamKey{}, param)
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
	param := r.Context().Value(ParamKey{}).(map[string]string)
	return param[key]
}