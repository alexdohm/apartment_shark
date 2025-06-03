package telegram

import (
	"net/http"
	"net/http/httptest"
)

// stub server that returns given status
func newStubServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}
