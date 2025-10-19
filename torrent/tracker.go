package torrent

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"time"

	"torrent-client/bencode"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

type TrackerResponse struct {
	Interval int
	Peers    []Peer
}

func RequestPeers(torrent *TorrentFile, peerID [20]byte, port uint16) (*TrackerResponse, error) {
	url, err := torrent.BuildTrackerURL(peerID, port)
	if err != nil {
		return nil, fmt.Errorf("failed to build tracker URL: %w", err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to contact tracker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned HTTP %d", resp.StatusCode)
	}

	// Read response body
	body := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %w", err)
	}

	// Decode bencoded response
	decoded, err := bencode.Decode(body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tracker response: %w", err)
	}

	responseDict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("tracker response is not a dictionary")
	}

	// Check for failure reason
	if failureReason, exists := responseDict["failure reason"]; exists {
		if reason, ok := failureReason.(string); ok {
			return nil, fmt.Errorf("tracker error: %s", reason)
		}
	}

	// Extract interval
	interval, ok := responseDict["interval"].(int)
	if !ok {
		return nil, fmt.Errorf("missing or invalid interval in tracker response")
	}

	// Extract peers
	peersData, ok := responseDict["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid peers in tracker response")
	}

	peers, err := parsePeers([]byte(peersData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse peers: %w", err)
	}

	return &TrackerResponse{
		Interval: interval,
		Peers:    peers,
	}, nil
}

func parsePeers(peersData []byte) ([]Peer, error) {
	const peerSize = 6 // 4 bytes IP + 2 bytes port

	if len(peersData)%peerSize != 0 {
		return nil, fmt.Errorf("invalid peers data length: %d (must be multiple of %d)", len(peersData), peerSize)
	}

	numPeers := len(peersData) / peerSize
	peers := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		offset := i * peerSize

		// Extract IP (4 bytes, network byte order)
		ip := net.IP(peersData[offset : offset+4])

		// Extract port (2 bytes, network byte order)
		port := binary.BigEndian.Uint16(peersData[offset+4 : offset+6])

		peers[i] = Peer{
			IP:   ip,
			Port: port,
		}
	}

	return peers, nil
}
