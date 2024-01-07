package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"
)

// Logging Middleware
type loggerMiddleware struct{
	next http.Handler
	*app
}
func (l loggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.next.ServeHTTP(w,r)
	l.logger.LogAttrs(context.Background(), slog.LevelInfo,"http request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("time_elapsed",time.Since(start).String()))
}
func (a *app) NewLoggerMiddleware(nextHandler http.Handler) loggerMiddleware {
	return loggerMiddleware{nextHandler, a}
}

// Auth Middleware
type authMiddleware struct{
	next http.Handler
	*app
}
type authKey struct {}
type sessionInfo struct {
	auth bool
	userId int
}
func (auth authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_cookie")
	if err != nil {
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	token := sessionCookie.Value
	userId, err := auth.getUserIdFromActiveSession(token)
	if err != nil {
		if err != sql.ErrNoRows {
			auth.logger.Error("erro ao acessar database", slog.String("errMsg",err.Error()))
		}
		ctx := context.WithValue(r.Context(), authKey{},sessionInfo{auth:false})
		auth.next.ServeHTTP(w,r.WithContext(ctx))
		return
	}
	ctx := context.WithValue(r.Context(), authKey{},sessionInfo{true,userId})
	auth.next.ServeHTTP(w,r.WithContext(ctx))
	// essa função escreve no db em TODA requisição de users
	// autenticados e na minha maquina adiciona 8-10ms a toda req,
	// será que devia fazer um cache?
	auth.prolongSession(token)
}

func (a *app)NewAuthMiddleware(nextHandler http.Handler) authMiddleware {
	return authMiddleware{nextHandler,a}
}

func (a *app)protected(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authInfo := r.Context().Value(authKey{}).(sessionInfo)
		if authInfo.auth {
			next(w,r)
			return
		}
		a.errorTmplHandler(w,"Não autorizado", http.StatusUnauthorized)
	}
}

// Error Handling Middleware
type errorMiddleware struct{
	next http.Handler
	*app
}
type errorResponseWrapper struct {
	http.ResponseWriter
	wasWritten bool
	status int
}
func (e *errorResponseWrapper) Write(bytes []byte) (int, error) {
	e.wasWritten = true
	return e.ResponseWriter.Write(bytes)
}
func (e *errorResponseWrapper) WriteHeader(status int) {
	e.status = status
	e.ResponseWriter.WriteHeader(status)
}
func (e errorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapped := &errorResponseWrapper{ResponseWriter: w}
	e.next.ServeHTTP(wrapped,r)
	if !wrapped.wasWritten && wrapped.status > 399 {
		e.errorTmplHandler(w,"",wrapped.status)
	}
}

func (a *app) NewErrorMiddleware(nextHandler http.Handler) errorMiddleware {
	return errorMiddleware{nextHandler, a}
}