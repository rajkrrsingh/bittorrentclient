package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"torrent-client/bencode"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func Open(path string) (*TorrentFile, error) {
	var data []byte
	var err error

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		data, err = downloadFile(path)
	} else {
		data, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read torrent file: %w", err)
	}

	return Parse(data)
}

func downloadFile(urlStr string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to download torrent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func Parse(data []byte) (*TorrentFile, error) {
	decoded, err := bencode.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode torrent: %w", err)
	}

	torrentDict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, errors.New("torrent file must be a dictionary")
	}

	torrent, err := parseTorrentDict(torrentDict)
	if err != nil {
		return nil, err
	}

	// Calculate info hash
	infoDict, ok := torrentDict["info"]
	if !ok {
		return nil, errors.New("missing info dictionary")
	}

	infoEncoded, err := bencode.Encode(infoDict)
	if err != nil {
		return nil, fmt.Errorf("failed to encode info dict: %w", err)
	}

	torrent.InfoHash = sha1.Sum(infoEncoded)

	return torrent, nil
}

func parseTorrentDict(dict map[string]interface{}) (*TorrentFile, error) {
	announce, ok := dict["announce"].(string)
	if !ok {
		return nil, errors.New("missing or invalid announce URL")
	}

	infoDict, ok := dict["info"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid info dictionary")
	}

	pieces, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, errors.New("missing or invalid pieces")
	}

	pieceLength, ok := infoDict["piece length"].(int)
	if !ok {
		return nil, errors.New("missing or invalid piece length")
	}

	length, ok := infoDict["length"].(int)
	if !ok {
		return nil, errors.New("missing or invalid length")
	}

	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, errors.New("missing or invalid name")
	}

	// Parse piece hashes
	if len(pieces)%20 != 0 {
		return nil, errors.New("invalid pieces length (must be multiple of 20)")
	}

	numPieces := len(pieces) / 20
	pieceHashes := make([][20]byte, numPieces)

	for i := 0; i < numPieces; i++ {
		copy(pieceHashes[i][:], pieces[i*20:(i+1)*20])
	}

	return &TorrentFile{
		Announce:    announce,
		PieceHashes: pieceHashes,
		PieceLength: pieceLength,
		Length:      length,
		Name:        name,
	}, nil
}

func (t *TorrentFile) BuildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}
