package torrent

import (
	"bytes"
	"crypto/sha1"
	"testing"

	"torrent-client/bencode"
)

func TestParseTorrent(t *testing.T) {
	// Create a sample torrent file data
	info := map[string]interface{}{
		"pieces":       string(bytes.Repeat([]byte("abcdefghij1234567890"), 2)), // 2 pieces of 20 bytes each
		"piece length": 262144,                                                  // 256KB
		"length":       524288,                                                  // 512KB total
		"name":         "test.txt",
	}

	torrentDict := map[string]interface{}{
		"announce": "http://tracker.example.com:8080/announce",
		"info":     info,
	}

	torrentData, err := bencode.Encode(torrentDict)
	if err != nil {
		t.Fatalf("Failed to encode test torrent: %v", err)
	}

	// Parse the torrent
	torrent, err := Parse(torrentData)
	if err != nil {
		t.Fatalf("Failed to parse torrent: %v", err)
	}

	// Verify fields
	if torrent.Announce != "http://tracker.example.com:8080/announce" {
		t.Errorf("Expected announce URL to be 'http://tracker.example.com:8080/announce', got '%s'", torrent.Announce)
	}

	if torrent.PieceLength != 262144 {
		t.Errorf("Expected piece length to be 262144, got %d", torrent.PieceLength)
	}

	if torrent.Length != 524288 {
		t.Errorf("Expected length to be 524288, got %d", torrent.Length)
	}

	if torrent.Name != "test.txt" {
		t.Errorf("Expected name to be 'test.txt', got '%s'", torrent.Name)
	}

	if len(torrent.PieceHashes) != 2 {
		t.Errorf("Expected 2 piece hashes, got %d", len(torrent.PieceHashes))
	}

	// Verify info hash calculation
	infoEncoded, err := bencode.Encode(info)
	if err != nil {
		t.Fatalf("Failed to encode info dict: %v", err)
	}
	expectedHash := sha1.Sum(infoEncoded)

	if torrent.InfoHash != expectedHash {
		t.Errorf("Info hash mismatch")
	}
}

func TestBuildTrackerURL(t *testing.T) {
	torrent := &TorrentFile{
		Announce: "http://tracker.example.com:8080/announce",
		InfoHash: [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		Length:   1024,
	}

	peerID := [20]byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T'}
	port := uint16(6881)

	url, err := torrent.BuildTrackerURL(peerID, port)
	if err != nil {
		t.Fatalf("Failed to build tracker URL: %v", err)
	}

	// Check that URL contains expected components
	expectedComponents := []string{
		"http://tracker.example.com:8080/announce",
		"port=6881",
		"uploaded=0",
		"downloaded=0",
		"compact=1",
		"left=1024",
	}

	for _, component := range expectedComponents {
		if !bytes.Contains([]byte(url), []byte(component)) {
			t.Errorf("URL '%s' missing expected component '%s'", url, component)
		}
	}
}

func TestParseErrors(t *testing.T) {
	testCases := []struct {
		name     string
		data     map[string]interface{}
		errorMsg string
	}{
		{
			name:     "missing announce",
			data:     map[string]interface{}{"info": map[string]interface{}{}},
			errorMsg: "missing or invalid announce URL",
		},
		{
			name:     "missing info",
			data:     map[string]interface{}{"announce": "http://example.com"},
			errorMsg: "missing or invalid info dictionary",
		},
		{
			name: "missing pieces",
			data: map[string]interface{}{
				"announce": "http://example.com",
				"info": map[string]interface{}{
					"piece length": 262144,
					"length":       1024,
					"name":         "test",
				},
			},
			errorMsg: "missing or invalid pieces",
		},
		{
			name: "invalid pieces length",
			data: map[string]interface{}{
				"announce": "http://example.com",
				"info": map[string]interface{}{
					"pieces":       "invalidlength", // Not multiple of 20
					"piece length": 262144,
					"length":       1024,
					"name":         "test",
				},
			},
			errorMsg: "invalid pieces length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := bencode.Encode(tc.data)
			if err != nil {
				t.Fatalf("Failed to encode test data: %v", err)
			}

			_, err = Parse(data)
			if err == nil {
				t.Errorf("Expected error containing '%s', but got nil", tc.errorMsg)
			} else if !bytes.Contains([]byte(err.Error()), []byte(tc.errorMsg)) {
				t.Errorf("Expected error containing '%s', got '%s'", tc.errorMsg, err.Error())
			}
		})
	}
}
