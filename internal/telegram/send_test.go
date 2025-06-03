package telegram

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestInit_Success(t *testing.T) {
	resetGlobalState()

	err := Init("telegram.io", "1234", "5678")
	if err != nil {
		t.Errorf("Init() expected no error, got %v", err)
	}
	if defaultNotifier == nil {
		t.Errorf("Init() expected defaultNotifer to be set but got nil")
	}
}

func TestInit_EmptyParameters(t *testing.T) {
	resetGlobalState()

	err := Init("", "", "")
	if err == nil {
		t.Errorf("Init() expected an error for empty paramterts")
	}
	if !strings.Contains(err.Error(), "must be non empty") {
		t.Errorf("Init() error message incorrect, got %v", err)
	}
	if defaultNotifier != nil {
		t.Errorf("Init() default notifier expected to be nil, but got value")
	}

}

func TestSend_NotInitialized(t *testing.T) {
	resetGlobalState()

	info := &TelegramInfo{
		Address:     "Test",
		Size:        "50",
		Rent:        "500",
		MapLink:     "some link",
		ListingLink: "some link",
	}

	err := Send(context.Background(), info, "test site")
	if err == nil {
		t.Errorf("Send() error expected but got nil")
	}
	if !strings.Contains(err.Error(), "not been initialized") {
		t.Errorf("Send() expected intialization error, got %v", err)
	}
}

func TestSend_Success(t *testing.T) {
	resetGlobalState()

	server := newStubServer(http.StatusOK)
	defer server.Close()

	err := Init(server.URL, "1234", "5678")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	info := &TelegramInfo{
		Address:     "Test",
		Size:        "50",
		Rent:        "500",
		MapLink:     "some link",
		ListingLink: "some link",
	}

	err = Send(context.Background(), info, "some site")
	if err != nil {
		t.Errorf("unexpected error sending: %v", err)
	}
}

func TestSend_HTTPError(t *testing.T) {
	resetGlobalState()

	server := newStubServer(http.StatusBadRequest)
	defer server.Close()

	err := Init(server.URL, "1234", "5678")
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	info := &TelegramInfo{
		Address:     "addy",
		Size:        "20",
		Rent:        "200",
		MapLink:     "someLink",
		ListingLink: "anotherLink",
	}
	err = Send(context.Background(), info, "TestSite")
	if err == nil {
		t.Error("Expected bad request but got nil")
	}

}

func resetGlobalState() {
	defaultNotifier = nil
	once = sync.Once{}
	initError = nil
}
