package telegram

import "fmt"

type TelegramInfo struct {
	Address, Size, Rent, MapLink, ListingLink string
}

func BuildHTML(info *TelegramInfo, site string) string {
	return fmt.Sprintf(`<b>%s Listing</b>

<b>Address:</b> %s
<b>Size:</b> %s m²
<b>Rent:</b> %s €

<a href="%s">View Map</a>

<a href="%s">View Listing</a>`,

		site, info.Address, info.Size, info.Rent, info.MapLink, info.ListingLink,
	)
}
