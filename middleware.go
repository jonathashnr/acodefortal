package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/jonathashnr/ajudafortaleza/database"
)

// ResponseWriter Wrapper
type responseWrapper struct {
	http.ResponseWriter
	status int
	writeHeaderCalled bool
	writeCalled bool
}

func (rw *responseWrapper) WriteHeader(code int) {
	if rw.writeHeaderCalled {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.writeHeaderCalled = true
}

func (rw *responseWrapper) Write(bytes []byte) (int, error) {
	if !rw.writeHeaderCalled {
		rw.WriteHeader(http.StatusOK)
	}
	rw.writeCalled = true
	return rw.ResponseWriter.Write(bytes)
}

func wrapResponseWriter(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{ResponseWriter: w}
}

// Logging and Error handling Middleware
type middleware struct{
	next http.Handler
	*app
}

func (m middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrapped := wrapResponseWriter(w)
	m.next.ServeHTTP(wrapped,r)
	m.logger.LogAttrs(context.Background(), slog.LevelInfo,"http request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("response",wrapped.status),slog.String("duration",time.Since(start).String()))
	if !wrapped.writeCalled && wrapped.status > 399 {
		m.errorTmplHandler(w,wrapped.status,"")
	}
}

func (a *app) NewMiddleware(nextHandler http.Handler) middleware {
	return middleware{nextHandler, a}
}

// Auth Middleware
type authMiddleware struct{
	next http.Handler
	*app
}

type authKey struct {}

type sessionInfo struct {
	auth bool
	user database.User
}

func (auth authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_cookie")
	if err != nil {
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	token := sessionCookie.Value
	user, err := auth.model.GetUserFromActiveSession(token)
	if err != nil {
		if err != sql.ErrNoRows {
			auth.logger.Error("erro ao acessar database", slog.String("errMsg",err.Error()))
		}
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	ctx := context.WithValue(r.Context(), authKey{},sessionInfo{true,user})
	auth.next.ServeHTTP(w,r.WithContext(ctx))
	// essa função escreve no db em TODA requisição de users
	// autenticados e na minha maquina adiciona 8-10ms a toda req,
	// será que devia fazer um cache?
	auth.model.ProlongSession(token)
}

func (a *app)NewAuthMiddleware(nextHandler http.Handler) authMiddleware {
	return authMiddleware{nextHandler,a}
}

func protected(next http.HandlerFunc, authzLevel int) http.HandlerFunc {
	// Níveis de acesso
	// 0    Usuários não-validados
	// 1    Usuários validados
	// >1   Níveis de permissão elevado (admins, superusers, etc.)
	return func(w http.ResponseWriter, r *http.Request) {
		authInfo := r.Context().Value(authKey{}).(sessionInfo)
		if !authInfo.auth {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if authzLevel > authInfo.user.Permission {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next(w,r)
	}
}