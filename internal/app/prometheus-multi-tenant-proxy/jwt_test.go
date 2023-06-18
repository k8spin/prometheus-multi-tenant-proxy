package proxy

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var (
	jwksHMAC = `{
		"keys": [
			{
				"kty": "oct",
				"kid": "hmac-key",
				"alg": "HS256",
				"k": "bGFsYQ=="
			}
		]
	}`

	// jwksJSON is a embedded JWKS in JSON format.
	//go:embed .jwks_example.json
	jwksJSON string
)

const (
	// kid = hmac-key, payload = {"namespaces": ["prometheus"]}
	validHmacToken = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImhtYWMta2V5In0.eyJuYW1lc3BhY2VzIjpbInByb21ldGhldXMiXX0.mGc9neZ2-C6fOXwI_h5Qknj-lH1apcFKVUo0-WlDPss"
	// rsa jwt geneator online: https://www.scottbrady91.com/tools/jwt
	// kid = "rs256-key", payload = {"namespaces": ["prometheus", "app-1"]}}
	validRsaToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6InJzMjU2LWtleSIsInR5cCI6IkpXVCJ9.eyJuYW1lc3BhY2VzIjpbInByb21ldGhldXMiLCJhcHAtMSJdfQ.n_hy5yqjFkpD00VNGCLkRyeOBdcjeu9Yp1TVzV5jSKaX32Idrl2jv1mHCX5JJfM-tyLXxCQJcze9q7IXpN0_x-E7iE_uAvDT7BiMWSwy7lWW2eRuffggv2EG8HP3_kGgsH-RcP4B5VbaKeB9N1RNrHwvxoiYKhcFQCTJzsf010s10nUYmfL0jQ8hW--yTX2kly8zXxBoJXu6rluNMXWL7o8Tx9ONHLLlz-trP7s9xFN_GQtbZ3lKZ5n8XESccctXWAdIqtYtlTA4KCr0krIX7cRbLdni5QOPBTwQxdOBujdDaXZqo8K8PJfaZ93oyJUdYe7rnX0Lz_dT1EJLWYvm-A"
	// kid = "rs256-key", payload = {"namespaces": []}}
	validEmptyNamespacesRSAToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6InJzMjU2LWtleSIsInR5cCI6IkpXVCJ9.eyJuYW1lc3BhY2VzIjpbXX0.LWPlxmZPDaKA3Z-IbRAoBYuymCx3cdZvXHzlSfVIhj4TjoQZ8Rom5IWtJpoEiq-_DkQHFgFRnLTsFE8CcaYM_eLWRZPK7c_rDwzfJDxDVhIL3k6krL5gq_4Y6nOGnjktJkIJvJstl9FDc7gyx0EBvUX-cgQzh-my9whMXBrZ0oybVyiBGlAZbVOiW-BObm3U0hYF4Xt6HOTm4khAEsZPnS4rglQpQki_q4w67OaMcTwfO_hr6KJtwzavLLCWJhijWdON93ueubn4Z294TM5SWQFzPM-knFDaBfzq5k94NQviBoT7ekb9RsGLrjKsrVOdOVMM8b4BEFXMtZpVENLgQg"
	// kid = "rs256-key", payload = {"labels": {"app":"ecom","team":"europe"}}}
	validTwoLabelsRSAToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6InJzMjU2LWtleSJ9.eyJsYWJlbHMiOnsiYXBwIjoiZWNvbSIsInRlYW0iOiJldXJvcGUifX0.Xb9-WPdP-yL7afsYShGQt1p3YVhNcufY-6dxtCVnbhgKLotqgy81tS-5RxF7KdSlSkfuwNyZCuE_qnKO_seOxczHOkARWnvZ5jlfPoPI8adKiVykeDR6q6fj3fO5Mp7BDNVXBwb9_wQ08Y3JwONdoNmvdnUz6aspD7IVIL41t64kst-GTxvvkdA-1Xfh9LB0zmyaCEgYiaByNJevtqnwFociTzRbWR2yXcEkhzbqKSGG6ia55It5CeN3GB9sjAWOEd57fSgDJwr0D80zxFoXtLeX64gcCjNsxJsh5ZrQ8U34fdo-73mPDJPCOBkowiamPDWOkBQ54U5lesbE5R3KPA"
	// kid = "rs256-key", payload = {"labels":{"app":"ecom","team":"europe"},"namespaces":["kube-system","monitoring"]}
	validTwoLabelsTwoNamespacesRSAToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6InJzMjU2LWtleSJ9.eyJsYWJlbHMiOnsiYXBwIjoiZWNvbSIsInRlYW0iOiJldXJvcGUifSwibmFtZXNwYWNlcyI6WyJrdWJlLXN5c3RlbSIsIm1vbml0b3JpbmciXX0.Zk6hE9OBUIH5ctMzSeq2p40dJFiwZS_TghePWlTB1_-XHOzZRGvbT-sXoZnIy1__lHJZ4h8t0-P0_zwQPpZ2aB2A0Ar3wogEiIdktoRtqQcMvSjjIjwNm8e9uaE1QBpeqNtxg5i3hDMJLVfsoXta0PJ9YW4hbuhnpaThhji9M7duOXv9eeW4nJHSFr3YVCn75qR35O8z3Pwjo_06OhSpK5sy1PbqQLNvzkWdKYiqAjezWnnh6kO37hQfDJWUaKxkhE4TmOMJRk_mRKrUpHZ1mQ6rZ4YXyo0pBBNqJ5uJYA45bT2FJpNqJ9rXHf2qjDBwcS6SEw8pDe-iRdIC0xr1Ig"
)

func (auth *JwtAuth) assertHmac(t *testing.T, expectAuthorized bool) {
	authorized, _, _ := auth.isAuthorized(validHmacToken)
	if authorized != expectAuthorized {
		t.Errorf("HMAC authorized=%v, expected=%v", authorized, expectAuthorized)
	}
}
func (auth *JwtAuth) assertRSA(t *testing.T, expectAuthorized bool) {
	authorized, _, _ := auth.isAuthorized(validRsaToken)
	if authorized != expectAuthorized {
		t.Errorf("RSA authorized=%v, expected=%v", authorized, expectAuthorized)
	}
}

func TestJWT_LoadFromURL(t *testing.T) {
	returnErr := false
	returnBody := jwksHMAC
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if returnErr {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnBody))
	}))
	defer server.Close()

	// Load only the HMAC key
	auth := NewJwtAuth(server.URL)
	if auth.isFile {
		t.Fatal("auth.isFile should be false")
	}
	auth.assertHmac(t, true)
	auth.assertRSA(t, false)

	// Reload and trigger and error, it should still work
	returnErr = true
	if auth.Load() {
		t.Error("The load should have failed")
	}
	auth.assertHmac(t, true)

	// Reload with all keys this time
	returnErr = false
	returnBody = jwksJSON
	if !auth.Load() {
		t.Error("The load should have succeeded")
	}
	auth.assertHmac(t, true)
	auth.assertRSA(t, true)
}

func TestJWT_LoadFromFile(t *testing.T) {
	file, err := os.CreateTemp("", "jwt_test")
	if err != nil {
		t.Fatalf("Could not create tempfile: %v", err)
	}
	defer os.Remove(file.Name())

	// Load only the HMAC key
	os.WriteFile(file.Name(), []byte(jwksHMAC), 0644)
	auth := NewJwtAuth(file.Name())
	if !auth.isFile {
		t.Fatal("auth.isFile should be false")
	}
	auth.assertHmac(t, true)
	auth.assertRSA(t, false)

	// Reload and trigger and error, it should still work
	os.WriteFile(file.Name(), []byte(""), 0644)
	if auth.Load() {
		t.Error("The load should have failed")
	}
	auth.assertHmac(t, true)
	auth.assertRSA(t, false)

	// Reload with all keys this time
	os.WriteFile(file.Name(), []byte(jwksJSON), 0644)
	if !auth.Load() {
		t.Error("The load should have succeeded")
	}
	auth.assertHmac(t, true)
	auth.assertRSA(t, true)
}

func TestJWT_IsAuthorized(t *testing.T) {
	auth := newJwtAuthFromString(jwksJSON)
	validTestCases := []struct {
		desc  string
		token string
		ns    []string
		l     map[string]string
	}{
		{"hmac", validHmacToken, []string{"prometheus"}, map[string]string{}},
		{"rsa", validRsaToken, []string{"prometheus", "app-1"}, map[string]string{}},
		{"empty-namespace-rsa", validEmptyNamespacesRSAToken, []string{}, map[string]string{}},
		{"two-labels-rsa", validTwoLabelsRSAToken, []string{}, map[string]string{"app": "ecom", "team": "europe"}},
		{"two-labels-two-namespaces-rsa", validTwoLabelsTwoNamespacesRSAToken, []string{"kube-system", "monitoring"}, map[string]string{"app": "ecom", "team": "europe"}},
	}

	for _, tc := range validTestCases {
		t.Run(tc.desc, func(t *testing.T) {
			authorized, namespaces, labels := auth.isAuthorized(tc.token)
			if !authorized {
				t.Fatal("Should be authorized")
			}
			if !reflect.DeepEqual(namespaces, tc.ns) {
				t.Fatalf("Got unexpected namespace: %v", namespaces)
			}
			if !reflect.DeepEqual(labels, tc.l) {
				t.Fatalf("Got unexpected labels: %v", labels)
			}
		})
	}

	invalidTestCases := []struct {
		reason string
		token  string
	}{
		{"empty", ""}, // Empty JWT.
		{"wrong key", "eyJhbGciOiJIUzI1NiIsImtpZCI6ImhtYWMta2V5In0.eyJuYW1lc3BhY2UiOiJwcm9tZXRoZXVzIn0.dY7Pwl4LLrBFkrK2krsYfj0PZdJSxHPSEtXGFozdhv0"},
		{"wrong kid", "eyJhbGciOiJIUzI1NiIsImtpZCI6InVua25vd24ifQ.eyJuYW1lc3BhY2UiOiJwcm9tZXRoZXVzIn0.IijHPJ7xExe_CTXJ0A1M9qwOCelnSuMkD8AV4JzvD8M"},
		{"claim wrong type", "eyJhbGciOiJIUzI1NiIsImtpZCI6ImhtYWMta2V5In0.eyJuYW1lc3BhY2VzIjp0cnVlfQ.oZNkqDopM6DVMADg-utHeAolMhfWmlUlxL88a9yOB0M"},
	}

	for _, tc := range invalidTestCases {
		t.Run(tc.reason, func(t *testing.T) {
			if authorized, _, _ := auth.isAuthorized(tc.token); authorized {
				t.Error("Signature should be invalid - invalid secret signature")
			}
		})
	}
}

func TestJWT_extractToken(t *testing.T) {
	testCases := []struct {
		desc      string
		setupFunc func(h *http.Header)
		expected  string
	}{
		{"empty", func(h *http.Header) {}, ""}, // Empty JWT.
		{"auth: valid", func(h *http.Header) { h.Add("Authorization", "Bearer someToken") }, "someToken"},
		{"auth: no bearer", func(h *http.Header) { h.Add("Authorization", "someToken") }, ""},
		{"auth: no token", func(h *http.Header) { h.Add("Authorization", "Bearer ") }, ""},
		{"auth: empty", func(h *http.Header) { h.Add("Authorization", "") }, ""},
		{"token: valid", func(h *http.Header) { h.Add("Token", "someToken") }, "someToken"},
		{"token: empty", func(h *http.Header) { h.Add("Token", "") }, ""},
		{"both", func(h *http.Header) { h.Add("Authorization", "Bearer auth"); h.Add("Token", "token") }, "auth"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			headers := &http.Header{}
			tc.setupFunc(headers)
			if token := extractTokens(headers); token != tc.expected {
				t.Errorf("Wrong token extracted: %v", token)
			}
		})
	}
}
