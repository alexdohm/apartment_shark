package listings

import (
	"log"
	"sync"
)

type ScraperState struct {
	mu       sync.Mutex
	listings map[string]bool
}

func NewScraperState() *ScraperState {
	return &ScraperState{
		listings: make(map[string]bool),
	}
}

// MarkAsSeen adds a listing ID to the map (thread-safe)
func (s *ScraperState) MarkAsSeen(postID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listings[postID] = true
}

// Exists checks if a listing ID has already been processed (thread-safe)
func (s *ScraperState) Exists(postID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listings[postID]
}

func (s *ScraperState) PrintListings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println("Listings:")
	for postID := range s.listings {
		log.Println(postID)
	}
}
