package howoge

type HowogeListing struct {
	ID      int     `json:"uid"`
	Address string  `json:"title"`
	Rent    float64 `json:"rent"`
	Size    float64 `json:"area"`
	Wbs     string  `json:"wbs"`
	Link    string  `json:"link"`
	Notice  string  `json:"notice"`
}

// HowogeResponse Struct for API response
type HowogeResponse struct {
	Results []HowogeListing `json:"immoobjects"`
}
