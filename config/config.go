package config

// telegram bot config
const (
	BotToken         = "7533179845:AAGS2FEsvPzyjwpdshzP2dTs3ctelCKcM80"
	ChatID           = "1392436626"
	TimeBetweenCalls = 20
)

// state urls
const (
	GewobagURL      = "https://www.gewobag.de/fuer-mieter-und-mietinteressenten/mietangebote/?objekttyp%5B%5D=wohnung&gesamtmiete_von=&gesamtmiete_bis=&gesamtflaeche_von=&gesamtflaeche_bis=&zimmer_von=&zimmer_bis=&sort-by="
	WbmURL          = "https://www.wbm.de/wohnungen-berlin/angebote/"
	HowogeURL       = "https://www.howoge.de/?type=999"
	DewegoURL       = "https://www.degewo.de/immosuche"
	StadtUndLandURL = "https://d2396ha8oiavw0.cloudfront.net/sul-main/immoSearch"
)

// search filters
var (
	ZipCodes = []string{
		"12043", "12045", "12047", "12049", "12051", "12053", "12055", "12059", // nk
		"10243", "10245", "10247", // fhain
		"10961", "10965", "10967", "10969", "10997", "10999", //xberg
		"12435", // alt-treptow
		"10179", // mitte
	}
	MinWarm = 600.0
	MaxWarm = 1000.0
	MinSqm  = 50.0
	Wbs     = false
)
