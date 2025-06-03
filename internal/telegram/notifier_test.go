package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stub server that returns given status
func newStubServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

// tiny mock that satisfies interface
type mockNotifier struct{ msgs []string }

func (m *mockNotifier) Send(_ context.Context, msg string) error {
	m.msgs = append(m.msgs, msg)
	return nil
}

func TestTelegramNotifier_OK(t *testing.T) {
	ts := newStubServer(http.StatusOK)
	defer ts.Close()
	cli := ts.Client()

	not := NewTelegramNotifier(ts.URL, "dummy-token", "1234", cli)

	ctx := context.Background()
	if err := not.Send(ctx, "hi"); err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
}

func TestTelegramNotifier_BadRequest(t *testing.T) {
	ts := newStubServer(http.StatusBadRequest)
	defer ts.Close()
	cli := ts.Client()

	not := NewTelegramNotifier(ts.URL, "dummy-token", "1234", cli)

	ctx := context.Background()
	if err := not.Send(ctx, "hi"); err == nil {
		t.Fatalf("want an err, but got nil")
	}
}

func TestTelegramNotifier_ContextCancelled(t *testing.T) {
	// simulate a slow server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // longer response than ctx timeout
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	cli := ts.Client()
	not := NewTelegramNotifier(ts.URL, "token", "chat", cli)

	err := not.Send(ctx, "test")
	if err == nil {
		t.Fatalf("expected error due to cancelled context")
	}
}
