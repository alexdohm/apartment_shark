package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func createTestClient(serverURL string) (*Client, error) {
	return NewClient(serverURL, "test-token", "test-chat-id")
}

// TestNewClient tests the Client constructor with validation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		botToken        string
		chatID          string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "valid parameters",
			baseURL:  "https://api.telegram.org",
			botToken: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			chatID:   "123456789",
			wantErr:  false,
		},
		{
			name:            "empty baseURL",
			baseURL:         "",
			botToken:        "valid-token",
			chatID:          "valid-chat",
			wantErr:         true,
			wantErrContains: "baseURL cannot be empty",
		},
		{
			name:            "empty botToken",
			baseURL:         "https://api.telegram.org",
			botToken:        "",
			chatID:          "valid-chat",
			wantErr:         true,
			wantErrContains: "botToken cannot be empty",
		},
		{
			name:            "empty chatID",
			baseURL:         "https://api.telegram.org",
			botToken:        "valid-token",
			chatID:          "",
			wantErr:         true,
			wantErrContains: "chatID cannot be empty",
		},
		{
			name:            "all empty parameters",
			baseURL:         "",
			botToken:        "",
			chatID:          "",
			wantErr:         true,
			wantErrContains: "baseURL cannot be empty", // First error encountered
		},
		{
			name:     "localhost URL",
			baseURL:  "http://localhost:8080",
			botToken: "test-token",
			chatID:   "test-chat",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.botToken, tt.chatID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() error = nil, want error")
				} else if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("NewClient() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
				}
				if client != nil {
					t.Errorf("NewClient() client = %v, want nil", client)
				}
			} else {
				if err != nil {
					t.Errorf("NewClient() unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("NewClient() returned nil client")
				}
				if client.BaseURL != tt.baseURL {
					t.Errorf("BaseURL = %v, want %v", client.BaseURL, tt.baseURL)
				}
				if client.BotToken != tt.botToken {
					t.Errorf("BotToken = %v, want %v", client.BotToken, tt.botToken)
				}
				if client.ChatID != tt.chatID {
					t.Errorf("ChatID = %v, want %v", client.ChatID, tt.chatID)
				}
				if client.HTTPClient == nil {
					t.Error("HTTPClient should not be nil")
				}
			}
		})
	}
}

// TestClient_SendMessage tests the core SendMessage functionality
func TestClient_SendMessage(t *testing.T) {
	tests := []struct {
		name            string
		serverStatus    int
		message         string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:         "successful send",
			serverStatus: http.StatusOK,
			message:      "Hello, World!",
			wantErr:      false,
		},
		{
			name:            "server returns bad request",
			serverStatus:    http.StatusBadRequest,
			message:         "test message",
			wantErr:         true,
			wantErrContains: "unexpected status",
		},
		{
			name:            "server returns internal error",
			serverStatus:    http.StatusInternalServerError,
			message:         "test message",
			wantErr:         true,
			wantErrContains: "unexpected status",
		},
		{
			name:         "empty message",
			serverStatus: http.StatusOK,
			message:      "",
			wantErr:      false,
		},
		{
			name:         "HTML message",
			serverStatus: http.StatusOK,
			message:      "<b>Bold</b> and <i>italic</i> text",
			wantErr:      false,
		},
		{
			name:         "long message",
			serverStatus: http.StatusOK,
			message:      strings.Repeat("Long message ", 100),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newStubServer(tt.serverStatus)
			defer server.Close()

			client, _ := createTestClient(server.URL)
			ctx := context.Background()

			err := client.SendMessage(ctx, tt.message)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SendMessage() error = nil, want error")
				} else if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("SendMessage() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("SendMessage() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestClient_SendMessage_ContextHandling tests context cancellation and timeout
func TestClient_SendMessage_ContextHandling(t *testing.T) {
	tests := []struct {
		name            string
		setupContext    func() (context.Context, context.CancelFunc)
		serverDelay     time.Duration
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "context timeout",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
			serverDelay:     50 * time.Millisecond, // Longer than timeout
			wantErr:         true,
			wantErrContains: "deadline exceeded",
		},
		{
			name: "context cancellation",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				// Cancel after a short delay
				go func() {
					time.Sleep(5 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			serverDelay:     20 * time.Millisecond,
			wantErr:         true,
			wantErrContains: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server with delay
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverDelay > 0 {
					time.Sleep(tt.serverDelay)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, _ := createTestClient(server.URL)
			ctx, cancel := tt.setupContext()
			defer cancel()

			err := client.SendMessage(ctx, "test message")

			if tt.wantErr {
				if err == nil {
					t.Errorf("SendMessage() error = nil, want error")
				} else if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("SendMessage() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("SendMessage() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestClient_SendListing tests the SendListing convenience method
func TestClient_SendListing(t *testing.T) {
	tests := []struct {
		name         string
		info         *TelegramInfo
		serverStatus int
		wantErr      bool
	}{
		{
			name: "complete listing info",
			info: &TelegramInfo{
				Address:     "Musterstraße 1, 10115 Berlin",
				Size:        "45.5",
				Rent:        "800",
				MapLink:     "https://maps.google.com/test",
				ListingLink: "https://example.com/listing/123",
				Site:        "TestSite",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty listing info",
			info: &TelegramInfo{
				Address:     "",
				Size:        "",
				Rent:        "",
				MapLink:     "",
				ListingLink: "",
				Site:        "",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "partial listing info",
			info: &TelegramInfo{
				Address: "Some Address",
				Size:    "50",
				Rent:    "900",
				// Missing links
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "server error",
			info: &TelegramInfo{
				Address: "Test Address",
				Size:    "40",
				Rent:    "700",
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "nil info",
			info:         nil,
			serverStatus: http.StatusOK,
			wantErr:      false, // BuildHTML should handle nil gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newStubServer(tt.serverStatus)
			defer server.Close()

			client, _ := createTestClient(server.URL)
			ctx := context.Background()

			err := client.SendListing(ctx, tt.info)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SendListing() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("SendListing() unexpected error: %v", err)
				}
			}
		})
	}
}
