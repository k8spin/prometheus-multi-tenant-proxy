package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type testAuth struct {
	authorized bool
	namespaces []string
	labels     map[string]string
	wasDenied  bool
}

func (a *testAuth) IsAuthorized(r *http.Request) (bool, []string, map[string]string) {
	return a.authorized, a.namespaces, a.labels
}

func (a *testAuth) WriteUnauthorisedResponse(w http.ResponseWriter) {
	a.wasDenied = true
}

func (a *testAuth) Load() bool {
	return true
}

func TestAuth_Ctx(t *testing.T) {
	ns := []string{"ns1", "ns2"}
	labels := map[string]string{"label1": "value1", "label2": "value2"}
	auth := &testAuth{
		authorized: true,
		namespaces: ns,
		labels:     labels,
	}

	r := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()
	h := func(w http.ResponseWriter, req *http.Request) {
		r = req
	}

	AuthHandler(auth, nil, h)(w, r)
	if auth.wasDenied {
		t.Errorf("Auth should be successful")
	}
	if !reflect.DeepEqual(ns, r.Context().Value(Namespaces).([]string)) {
		t.Errorf("Namespaces should be set")
	}
	if !reflect.DeepEqual(labels, r.Context().Value(Labels).(map[string]string)) {
		t.Errorf("Labels should be set")
	}
}

func TestAuth_Whitelist(t *testing.T) {
	auth := &testAuth{
		authorized: true,
		namespaces: []string{"ns"},
		labels:     map[string]string{},
	}
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	h := func(w http.ResponseWriter, req *http.Request) {}

	testCases := []struct {
		whitelist []string
		ok        bool
	}{
		{nil, true},
		{[]string{}, false},
		{[]string{"/foo"}, true},
		{[]string{"/bar"}, false},
		{[]string{"/bar/foo"}, false},
		{[]string{"/foo/bar"}, false},
		{[]string{"/bar", "/buzz", "/foo"}, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.whitelist), func(t *testing.T) {
			auth.wasDenied = false // reset
			AuthHandler(auth, tc.whitelist, h)(httptest.NewRecorder(), r)
			if auth.wasDenied == tc.ok {
				t.Errorf("Whitelist %v should return ok=%v", tc.whitelist, tc.ok)
			}
		})
	}
}

func TestAuth_AuthHandler(t *testing.T) {
	ns := []string{"ns1"}
	ls := map[string]string{"foo": "bar"}
	noNs := []string{}
	noLs := map[string]string{}

	testCases := []struct {
		authorized bool
		ns         []string
		ls         map[string]string
		ok         bool
	}{
		{true, ns, ls, true},
		{true, ns, noLs, true},
		{true, noNs, ls, true},
		{true, noNs, noLs, false},
		{false, ns, ls, false},
		{false, noNs, noLs, false},
	}

	for _, tc := range testCases {
		desc := fmt.Sprintf("A=%v,N=%v,L=%v", tc.authorized, len(tc.ns) == 0, len(tc.ls) == 0)
		t.Run(desc, func(t *testing.T) {
			auth := &testAuth{
				authorized: tc.authorized,
				namespaces: tc.ns,
				labels:     tc.ls,
				wasDenied:  false,
			}

			handlerCalled := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
			}

			r := httptest.NewRequest("GET", "http://example.com", nil)
			w := httptest.NewRecorder()

			AuthHandler(auth, nil, handler)(w, r)
			if tc.ok != handlerCalled {
				t.Errorf("handler called: %v should have been %v", handlerCalled, tc.ok)
			}
			if tc.ok == auth.wasDenied {
				t.Errorf("denied: %v should have been %v", handlerCalled, tc.ok)
			}
		})
	}
}

func TestAuth_isInWhitelist(t *testing.T) {
	whitelist := []string{
		"/api/v1/query",
		"/foo",
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{"", false},
		{"/foo", true},
		{"/bar/foo", true},
		{"/foo/bar", false},
		{"/api/v1/query", true},
		{"/v1/query", false},
		{"/api/v2/query", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			if isInWhitelist("https://example.com"+tc.path, whitelist) != tc.expected {
				t.Errorf("%s != %v", tc.path, tc.expected)
			}
			if isInWhitelist("https://example.com/some-path"+tc.path, whitelist) != tc.expected {
				t.Errorf("%s != %v", tc.path, tc.expected)
			}
		})
	}
}
