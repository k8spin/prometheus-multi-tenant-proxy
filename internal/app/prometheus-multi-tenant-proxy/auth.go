package proxy

import (
	"context"
	"crypto/subtle"
	"net/http"
)

type key int

const (
	//Namespace Key used to pass prometheus tenant id though the middleware context
	Namespace key = iota
	realm         = "Prometheus multi-tenant proxy"
)

// BasicAuth can be used as a middleware chain to authenticate users before proxying a request
func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		authorized, namespace := isAuthorized(user, pass)
		if !ok || !authorized {
			writeUnauthorisedResponse(w)
			return
		}
		ctx := context.WithValue(r.Context(), Namespace, namespace)
		handler(w, r.WithContext(ctx))
	}
}

func isAuthorized(user string, pass string) (bool, string) {
	authConfig := GetConfig()
	for _, v := range authConfig.Users {
		if subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(v.Password)) == 1 {
			return true, v.Namespace
		}
	}
	return false, ""
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}
