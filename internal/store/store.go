package store

import (
	"log"
	"sync"
)

type ScraperStore interface {
	MarkAsSeen(string)
	Exists(string) bool
	PrintListings()
	Size() int
	Clear()
}

type ScraperState struct {
	mu       sync.RWMutex
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

func (s *ScraperState) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.listings)
}

func (s *ScraperState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listings = make(map[string]bool)
}
