package torrent

import (
	"encoding/binary"
	"net"
	"testing"

	"torrent-client/bencode"
)

func TestParsePeers(t *testing.T) {
	// Create test peer data: 2 peers
	// Peer 1: 192.168.1.1:8080
	// Peer 2: 10.0.0.1:6881
	peersData := make([]byte, 12) // 2 peers * 6 bytes each

	// Peer 1: 192.168.1.1:8080
	peersData[0] = 192
	peersData[1] = 168
	peersData[2] = 1
	peersData[3] = 1
	binary.BigEndian.PutUint16(peersData[4:6], 8080)

	// Peer 2: 10.0.0.1:6881
	peersData[6] = 10
	peersData[7] = 0
	peersData[8] = 0
	peersData[9] = 1
	binary.BigEndian.PutUint16(peersData[10:12], 6881)

	peers, err := parsePeers(peersData)
	if err != nil {
		t.Fatalf("parsePeers failed: %v", err)
	}

	if len(peers) != 2 {
		t.Fatalf("Expected 2 peers, got %d", len(peers))
	}

	// Check first peer
	expectedIP1 := net.IPv4(192, 168, 1, 1)
	if !peers[0].IP.Equal(expectedIP1) {
		t.Errorf("Peer 0 IP: expected %s, got %s", expectedIP1, peers[0].IP)
	}
	if peers[0].Port != 8080 {
		t.Errorf("Peer 0 Port: expected 8080, got %d", peers[0].Port)
	}

	// Check second peer
	expectedIP2 := net.IPv4(10, 0, 0, 1)
	if !peers[1].IP.Equal(expectedIP2) {
		t.Errorf("Peer 1 IP: expected %s, got %s", expectedIP2, peers[1].IP)
	}
	if peers[1].Port != 6881 {
		t.Errorf("Peer 1 Port: expected 6881, got %d", peers[1].Port)
	}
}

func TestParsePeersInvalidLength(t *testing.T) {
	// Invalid length (not multiple of 6)
	invalidData := make([]byte, 5)

	_, err := parsePeers(invalidData)
	if err == nil {
		t.Error("Expected error for invalid peer data length, got nil")
	}
}

func TestPeerString(t *testing.T) {
	peer := Peer{
		IP:   net.IPv4(192, 168, 1, 100),
		Port: 6881,
	}

	expected := "192.168.1.100:6881"
	if peer.String() != expected {
		t.Errorf("Peer.String(): expected %s, got %s", expected, peer.String())
	}
}

func TestTrackerResponseParsing(t *testing.T) {
	// Create mock tracker response data
	peersData := make([]byte, 6) // 1 peer
	peersData[0] = 127
	peersData[1] = 0
	peersData[2] = 0
	peersData[3] = 1
	binary.BigEndian.PutUint16(peersData[4:6], 8080)

	responseDict := map[string]interface{}{
		"interval": 1800, // 30 minutes
		"peers":    string(peersData),
	}

	responseData, err := bencode.Encode(responseDict)
	if err != nil {
		t.Fatalf("Failed to encode mock response: %v", err)
	}

	// Decode the response (simulating what happens in RequestPeers)
	decoded, err := bencode.Decode(responseData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	responseMap := decoded.(map[string]interface{})

	// Extract interval
	interval, ok := responseMap["interval"].(int)
	if !ok {
		t.Fatal("Missing or invalid interval")
	}
	if interval != 1800 {
		t.Errorf("Expected interval 1800, got %d", interval)
	}

	// Extract and parse peers
	peersStr, ok := responseMap["peers"].(string)
	if !ok {
		t.Fatal("Missing or invalid peers")
	}

	peers, err := parsePeers([]byte(peersStr))
	if err != nil {
		t.Fatalf("Failed to parse peers: %v", err)
	}

	if len(peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(peers))
	}

	expectedIP := net.IPv4(127, 0, 0, 1)
	if !peers[0].IP.Equal(expectedIP) {
		t.Errorf("Expected IP %s, got %s", expectedIP, peers[0].IP)
	}

	if peers[0].Port != 8080 {
		t.Errorf("Expected port 8080, got %d", peers[0].Port)
	}
}

func TestTrackerErrorResponse(t *testing.T) {
	// Create mock error response
	errorResponse := map[string]interface{}{
		"failure reason": "Invalid info_hash",
	}

	responseData, err := bencode.Encode(errorResponse)
	if err != nil {
		t.Fatalf("Failed to encode error response: %v", err)
	}

	// Decode and check for error
	decoded, err := bencode.Decode(responseData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	responseMap := decoded.(map[string]interface{})

	if failureReason, exists := responseMap["failure reason"]; exists {
		if reason, ok := failureReason.(string); ok {
			expectedError := "Invalid info_hash"
			if reason != expectedError {
				t.Errorf("Expected error '%s', got '%s'", expectedError, reason)
			}
		} else {
			t.Error("Failure reason is not a string")
		}
	} else {
		t.Error("No failure reason found in error response")
	}
}
