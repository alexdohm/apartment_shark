package telegram

import (
	"apartmenthunter/config"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type TelegramInfo struct {
	Address     string
	Size        string
	Rent        string
	MapLink     string
	ListingLink string
}

func GenerateTelegramMessage(info *TelegramInfo, site string) {
	htmlMsg := fmt.Sprintf(`<b>%s Listing</b>

<b>Address:</b> %s
<b>Size:</b> %s m²
<b>Rent:</b> %s €

<a href="%s">View Map</a>

<a href="%s">View Listing</a>`,

		site, info.Address, info.Size, info.Rent, info.MapLink, info.ListingLink,
	)
	SendTelegramMessage(htmlMsg)
}

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
