package bot

import (
	"math/rand"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Mobile Safari/537.36",
}

type HeaderGenerator struct{}

func NewHeaderGenerator() *HeaderGenerator {
	return &HeaderGenerator{}
}

func (h *HeaderGenerator) getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// GenerateGeneralRequestHeaders generate headers to help evade bot checks
// only set origin if cross-origin post request, or ajax request
func (h *HeaderGenerator) GenerateGeneralRequestHeaders(origin, referer string, formEncodedPost bool, jsonPost bool) map[string]string {
	headers := make(map[string]string)
	headers["User-Agent"] = h.getRandomUserAgent()
	if origin != "" {
		headers["Origin"] = origin
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	if formEncodedPost {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	if jsonPost {
		headers["Content-Type"] = "application/json"
		headers["Accept"] = "application/json"
	}
	return headers
}
