package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)

}

func (app *application) addIpToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.Background()
		ip, err := getIp(r)
		if err != nil {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if len(ip) == 0 {
				ip = "unknown"
			}
		}
		ctx = context.WithValue(r.Context(), contextUserKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIp(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown", err
	}
	userIp := net.ParseIP(ip)
	if userIp == nil {
		return "", fmt.Errorf("user IP: %q is not IP:port", r.RemoteAddr)
	}
	forward := r.Header.Get("X-Forwarded-for")
	if len(forward) > 0 {
		ip = forward
	}
	if len(ip) == 0 {
		ip = "forward"
	}
	return ip, nil
}

func (app *application) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user") {
			app.Session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
