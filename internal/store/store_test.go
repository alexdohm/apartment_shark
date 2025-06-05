package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewScraperState(t *testing.T) {
	state := NewScraperState()
	if state == nil {
		t.Fatal("scraper state failed to intialize")
	}
	if state.Size() != 0 {
		t.Errorf("Scraper state listings should be empty")
	}
}

func TestMarkAsSeenAndExists(t *testing.T) {
	state := NewScraperState()
	post := "1234"

	if state.Exists(post) {
		t.Errorf("post %s should not exist initially in the store", post)
	}
	state.MarkAsSeen(post)
	if !state.Exists(post) {
		t.Errorf("post %s should exist after marking as seen", post)
	}
	if state.Size() > 1 {
		t.Errorf("state size should be 1, got %d", state.Size())
	}
}

func TestMultipleDifferentPosts(t *testing.T) {
	state := NewScraperState()
	posts := []string{"post 1", "post 2", "post 3", "post 4"}

	for _, p := range posts {
		state.MarkAsSeen(p)
	}

	if state.Size() != len(posts) {
		t.Errorf("state size should be 4, got %d", state.Size())
	}
	for _, p := range posts {
		if !state.Exists(p) {
			t.Errorf("post %s should exist in the state", p)
		}
	}
}

func TestDuplicatePosts(t *testing.T) {
	state := NewScraperState()
	post := "post"
	state.MarkAsSeen(post)
	state.MarkAsSeen(post)
	state.MarkAsSeen(post)

	if state.Size() != 1 {
		t.Errorf("size of store should be 1, got %d", state.Size())
	}
	if !state.Exists(post) {
		t.Errorf("post %s should exist in store, but does not", post)
	}
}

func TestNonExistentPosts(t *testing.T) {
	state := NewScraperState()
	postExisting := "post"
	postNotExisting := "lost"

	state.MarkAsSeen(postExisting)
	if state.Exists(postNotExisting) {
		t.Errorf("post %s should not exist, but does", postNotExisting)
	}
}

func TestEdgeCases(t *testing.T) {
	state := NewScraperState()
	edgeCases := []string{
		"",
		" ",
		"very-long-post-id-with-lots-of-characters-and-numbers-12345",
		"post with spaces",
		"post/with/slashes",
		"post-with-special-chars!@#$%^&*()",
	}

	for _, p := range edgeCases {
		if state.Exists(p) {
			t.Errorf("post %s should not exist initially, but does", p)
		}
		state.MarkAsSeen(p)
		if !state.Exists(p) {
			t.Errorf("post %s should exist after insertion, but does not", p)
		}
	}

	if state.Size() != len(edgeCases) {
		t.Errorf("state size should be %d, but got %d", len(edgeCases), state.Size())
	}
}

func TestClear(t *testing.T) {
	state := NewScraperState()
	posts := []string{"post 1", "post 2", "post 3", "post 4"}

	for _, p := range posts {
		state.MarkAsSeen(p)
	}

	if state.Size() != len(posts) {
		t.Errorf("state size should be %d before clearing, but got %d", len(posts), state.Size())
	}
	state.Clear()
	if state.Size() != 0 {
		t.Errorf("store size should be 0 after clearing, got  %d", state.Size())
	}
}

func TestConcurrentAccess(t *testing.T) {
	state := NewScraperState()
	numGoRoutines := 10
	postsPerGoRoutine := 100

	var wg sync.WaitGroup

	// marking
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < postsPerGoRoutine; j++ {
				postID := fmt.Sprintf("post-%d-%d", id, j)
				state.MarkAsSeen(postID)
			}
		}(i)
	}

	//checking
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < postsPerGoRoutine; j++ {
				postID := fmt.Sprintf("post-%d-%d", id, j)
				state.Exists(postID) // dont care about viewing results in race condition testing
			}
		}(i)
	}

	wg.Wait()
	expectedSize := numGoRoutines * postsPerGoRoutine
	if state.Size() != expectedSize {
		t.Errorf("state should have size %d, but has size %d", expectedSize, state.Size())
	}
}

func BenchmarkConcurrentOperations(b *testing.B) {
	state := NewScraperState()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			postID := fmt.Sprintf("post-%d", i)
			if i%2 == 0 {
				state.MarkAsSeen(postID)
			} else {
				state.Exists(postID)
			}
			i++
		}
	})
}
