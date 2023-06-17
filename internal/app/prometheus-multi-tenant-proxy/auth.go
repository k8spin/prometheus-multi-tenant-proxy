package proxy

import (
	"context"
	"net/http"
)

type key int

// Auth implements an authentication middleware
type Auth interface {
	// IsAuthorized authenticates a request and returns the list of namespaces the user has access to
	IsAuthorized(r *http.Request) (bool, []string, map[string]string)
	// WriteUnauthorisedResponse writes an HTTP response in case the user is forbidden
	WriteUnauthorisedResponse(w http.ResponseWriter)
	// Load loads or reloads the configuration
	Load() bool
}

// AuthHandler returns au authentication middleware handler
func AuthHandler(auth Auth, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorized, namespaces, labels := auth.IsAuthorized(r)
		if !authorized {
			auth.WriteUnauthorisedResponse(w)
			return
		}
		ctx := context.WithValue(r.Context(), Namespaces, namespaces)
		ctx = context.WithValue(ctx, Labels, labels)
		handler(w, r.WithContext(ctx))
	}
}
