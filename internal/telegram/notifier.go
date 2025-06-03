package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Notifier interface {
	Send(context.Context, string) error
}

type TelegramNotifier struct {
	BaseURL  string
	BotToken string
	ChatID   string
	Client   *http.Client
}

func NewTelegramNotifier(base string, token string, id string, c *http.Client) *TelegramNotifier {
	return &TelegramNotifier{
		BaseURL:  base,
		BotToken: token,
		ChatID:   id,
		Client:   c,
	}
}

func (t *TelegramNotifier) Send(ctx context.Context, htmlMessage string) error {
	if t.Client == nil {
		t.Client = http.DefaultClient
	}
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", t.BaseURL, t.BotToken)
	formData := url.Values{
		"chat_id":                  {t.ChatID},
		"text":                     {htmlMessage},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %s", resp.Status)
	}
	return nil
}
