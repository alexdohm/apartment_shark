package bot

import (
	"apartmenthunter/config"
	"math/rand"
	"time"
)

// GenerateRandomJitterTime Add some randomness to the time between calls +- 30 seconds
func GenerateRandomJitterTime() time.Duration {
	return time.Duration(rand.Intn(config.TimeBetweenCalls)+30) * time.Second
}
