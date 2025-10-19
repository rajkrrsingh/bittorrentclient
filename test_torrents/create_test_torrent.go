// Test script to create a simple torrent file for testing
package main

import (
	"crypto/sha1"
	"fmt"
	"os"

	"torrent-client/bencode"
)

func main() {
	// Create a simple test file content
	testContent := "Hello, BitTorrent! This is a test file for our BitTorrent client."
	pieceLength := 32 // Very small for testing

	// Calculate pieces hashes
	var pieces []byte
	for i := 0; i < len(testContent); i += pieceLength {
		end := i + pieceLength
		if end > len(testContent) {
			end = len(testContent)
		}
		piece := testContent[i:end]
		hash := sha1.Sum([]byte(piece))
		pieces = append(pieces, hash[:]...)
	}

	// Create torrent structure
	info := map[string]interface{}{
		"name":         "test-file.txt",
		"length":       len(testContent),
		"piece length": pieceLength,
		"pieces":       string(pieces),
	}

	torrent := map[string]interface{}{
		"announce": "http://tracker.example.com:8080/announce",
		"info":     info,
	}

	// Encode to bencoded format
	data, err := bencode.Encode(torrent)
	if err != nil {
		fmt.Printf("Failed to encode torrent: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("test.torrent", data, 0644)
	if err != nil {
		fmt.Printf("Failed to write torrent file: %v\n", err)
		os.Exit(1)
	}

	// Also create the actual file content
	err = os.WriteFile("test-file.txt", []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("Failed to write test file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created test.torrent and test-file.txt")
	fmt.Printf("File size: %d bytes\n", len(testContent))
	fmt.Printf("Piece length: %d bytes\n", pieceLength)
	fmt.Printf("Number of pieces: %d\n", len(pieces)/20)
}
