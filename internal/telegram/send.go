package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	defaultNotifier Notifier
	once            sync.Once
	initError       error
)

func Init(base string, token string, chatID string) error {
	once.Do(func() {
		if base == "" || token == "" || chatID == "" {
			initError = errors.New("telegram: token, baseURL, chatID must be non empty")
			return
		}
		defaultNotifier = NewTelegramNotifier(base, token, chatID, &http.Client{Timeout: 5 * time.Second})
	})
	return initError
}

func Send(ctx context.Context, info *TelegramInfo) error {
	if defaultNotifier == nil {
		return fmt.Errorf("default notifier has not been initialized")
	}
	msg := BuildHTML(info)
	log.Println(msg)

	return defaultNotifier.Send(ctx, msg)
}

func SendStartup(ctx context.Context, msg string) error {
	return defaultNotifier.Send(ctx, msg)
}
