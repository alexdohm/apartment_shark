package telegram

import (
	"apartmenthunter/config"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func SendTelegramMessage(htmlMessage string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)

	formData := url.Values{
		"chat_id":                  {config.ChatID},
		"text":                     {htmlMessage},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
	}

	resp, err := http.PostForm(apiURL, formData)
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
		return
	}
	defer resp.Body.Close()
}
