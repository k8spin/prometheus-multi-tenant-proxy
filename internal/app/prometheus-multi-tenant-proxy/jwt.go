package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// NamespaceClaim expected structure of the JWT token payload
type NamespaceClaim struct {
	// Namespaces contains the list of namespaces a user has access to
	Namespaces []string `json:"namespaces"`
	jwt.RegisteredClaims
}

// JwtAuth can be used as a middleware chain to authenticate users
// using a JWT token before proxying a request
type JwtAuth struct {
	config     string
	isFile     bool
	b64content string
	jwks       *keyfunc.JWKS
	lock       *sync.RWMutex
}

// NewJwtAuth creates a JwtAuth by loaded a JWKS from either a file or an URL
func NewJwtAuth(config string) *JwtAuth {
	auth := &JwtAuth{
		config: config,
		isFile: true,
		lock:   new(sync.RWMutex),
	}
	if strings.HasPrefix(config, "http://") || strings.HasPrefix(config, "https://") {
		// We have a URL
		auth.isFile = false
	}
	if !auth.Load() {
		log.Fatal("Could not initialize JWT authentication.")
	}
	return auth
}

func newJwtAuthFromString(jwksJSON string) *JwtAuth {
	// ! the Load() method cannot be used.
	jwks, err := keyfunc.NewJSON(json.RawMessage(jwksJSON))
	if err != nil {
		log.Fatalf("Could not load JWKS: %v", err)
	}
	return &JwtAuth{
		jwks: jwks,
		lock: new(sync.RWMutex),
	}
}

func (auth *JwtAuth) String() string {
	s := fmt.Sprintf("JwtAuth{config: %s", auth.config)
	if auth.jwks != nil {
		s += fmt.Sprintf(", KIDs: %v", auth.jwks.KIDs())
	}
	s += "}"
	return s
}

// Load loads or reloads the JWKS from its config location (file or URL).
func (auth *JwtAuth) Load() bool {
	if auth.config == "" {
		log.Fatalf("JWTAuth: Load() cannot be called without a config")
	}

	if auth.isFile {
		return auth.loadFromFile(&auth.config)
	}
	return auth.loadFromURL(&auth.config)

}

func (auth *JwtAuth) loadFromURL(url *string) bool {
	// We do not use the jwks.Reload() method here
	// to avoid getting the lock unless strictly necessary.
	jwks, err := keyfunc.Get(*url, keyfunc.Options{})
	if err != nil {
		log.Printf("Failed to get the JWKS from the given URL: %s", err)
		return false
	}
	b64content := base64.StdEncoding.EncodeToString(jwks.RawJWKS())

	if auth.jwks == nil || b64content != auth.b64content {
		auth.lock.RLock()
		defer auth.lock.RUnlock()
		auth.jwks = jwks
		auth.b64content = b64content
		log.Printf("Reloaded JWKS from URL: %s", *url)
	}
	return true
}

func (auth *JwtAuth) loadFromFile(location *string) bool {
	content, err := ioutil.ReadFile(*location)
	if err != nil {
		log.Printf("Failed to read JWKS file: %v", err)
		return false
	}
	b64content := base64.StdEncoding.EncodeToString(content)
	if auth.b64content == b64content {
		// nothing to do
		return true
	}

	jwks, err := keyfunc.NewJSON(json.RawMessage(content))
	if err != nil {
		log.Printf("Failed to parse JWKS file: %v", err)
		return false
	}
	auth.lock.RLock()
	defer auth.lock.RUnlock()
	auth.b64content = b64content
	auth.jwks = jwks
	log.Print("Reloaded JWKS from file")
	return true
}

// IsAuthorized validates the user by verifying the JWT token in
// the request and returning the namespaces claim found in token the payload.
func (auth *JwtAuth) IsAuthorized(r *http.Request) (bool, []string) {
	tokenString := extractTokens(&r.Header)
	if tokenString == "" {
		log.Printf("Token is missing from header request")
		return false, nil
	}
	return auth.isAuthorized(tokenString)
}

// WriteUnauthorisedResponse writes a 401 Unauthorized HTTP response
func (auth *JwtAuth) WriteUnauthorisedResponse(w http.ResponseWriter) {
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}

func (auth *JwtAuth) isAuthorized(tokenString string) (bool, []string) {
	token, err := jwt.ParseWithClaims(tokenString, &NamespaceClaim{}, auth.jwks.Keyfunc)
	if err != nil || !token.Valid {
		log.Printf("%s\n", err)
		return false, nil
	}

	claims := token.Claims.(*NamespaceClaim)
	if len(claims.Namespaces) == 0 {
		log.Printf("token claim is invalid: namespaces is missing or empty")
		return false, nil
	}
	return true, claims.Namespaces
}

func isValidSigningMethod(signingMethod string) bool {
	for _, alg := range jwt.GetAlgorithms() {
		if signingMethod == alg {
			return true
		}
	}
	return false
}

func extractTokens(headers *http.Header) string {
	if token := headers.Get("Authorization"); token != "" {
		split := strings.Split(token, "Bearer ")
		if len(split) == 2 {
			return split[1]
		}
	}
	if token := headers.Get("Token"); token != "" {
		return token
	}
	return ""
}
