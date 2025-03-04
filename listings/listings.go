package listings

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// AppendListing writes a new line with the postID to the file
func AppendListing(filename, postID string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(postID + "\n"); err != nil {
		log.Printf("Failed to write postID to file: %v", err)
	}
}

// LoadListings loads seen post IDs into a map
func LoadListings(filename string) (map[string]bool, error) {
	m := make(map[string]bool)

	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		//no file, so return empty map
		return m, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			//mark the listing as being seen
			m[line] = true
		}
	}
	return m, scanner.Err()
}
