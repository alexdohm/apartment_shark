package stadtundland

type Address struct {
	Street      string `json:"street"`
	HouseNumber string `json:"house_number"`
	PostalCode  string `json:"postal_code"`
	City        string `json:"city"`
}

type Details struct {
	Id   string `json:"immoNumber"`
	Area string `json:"livingSpace"`
}

type Costs struct {
	Rent string `json:"warmRent"`
}

type StadtUndLandListing struct {
	Title   string  `json:"headline"`
	Address Address `json:"address"`
	Details Details `json:"details"`
	Costs   Costs   `json:"costs"`
	Link    string  `json:"url"`
}

type StadtUndLandResponse struct {
	Listings []StadtUndLandListing `json:"data"`
}
