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

	AuthHandler(auth, h)(w, r)
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

			AuthHandler(auth, handler)(w, r)
			if tc.ok != handlerCalled {
				t.Errorf("handler called: %v should have been %v", handlerCalled, tc.ok)
			}
			if tc.ok == auth.wasDenied {
				t.Errorf("denied: %v should have been %v", handlerCalled, tc.ok)
			}
		})
	}
}
