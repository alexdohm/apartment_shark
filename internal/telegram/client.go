package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles Telegram message sending
type Client struct {
	BaseURL    string
	BotToken   string
	ChatID     string
	HTTPClient *http.Client
}

// NewClient creates a new Telegram client
func NewClient(baseURL, botToken, chatID string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if botToken == "" {
		return nil, fmt.Errorf("botToken cannot be empty")
	}
	if chatID == "" {
		return nil, fmt.Errorf("chatID cannot be empty")
	}

	return &Client{
		BaseURL:    baseURL,
		BotToken:   botToken,
		ChatID:     chatID,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// SendMessage sends a raw HTML message
func (c *Client) SendMessage(ctx context.Context, htmlMessage string) error {
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", c.BaseURL, c.BotToken)
	formData := url.Values{
		"chat_id":                  {c.ChatID},
		"text":                     {htmlMessage},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %s", resp.Status)
	}
	return nil
}

// SendListing sends an apartment listing (convenience method)
func (c *Client) SendListing(ctx context.Context, info *TelegramInfo) error {
	message := BuildHTML(info)
	return c.SendMessage(ctx, message)
}

// SendStartup sends a startup notification
func (c *Client) SendStartup(ctx context.Context, message string) error {
	return c.SendMessage(ctx, message)
}
