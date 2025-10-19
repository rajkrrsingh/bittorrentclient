package client

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"torrent-client/peer"
	"torrent-client/torrent"
)

const Port uint16 = 6881

type Torrent struct {
	Peers       []torrent.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *peer.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

const MaxBlockSize = 16384
const MaxBacklog = 5

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() // this call blocks
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case peer.MsgUnchoke:
		state.client.Choked = false
	case peer.MsgChoke:
		state.client.Choked = true
	case peer.MsgHave:
		index, err := peer.ParseHaveMessage(msg.Payload)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case peer.MsgPiece:
		n, err := state.copyPieceData(state.index, msg.Payload)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func (state *pieceProgress) copyPieceData(index int, buf []byte) (int, error) {
	begin, block, err := peer.ParsePieceMessage(index, buf)
	if err != nil {
		return 0, err
	}

	if begin >= len(state.buf) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(state.buf))
	}

	if begin+len(block) > len(state.buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(block), begin, len(state.buf))
	}

	copy(state.buf[begin:], block)
	return len(block), nil
}

func attemptDownloadPiece(c *peer.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peerAddr torrent.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	peerStruct := &peer.Peer{IP: peerAddr.IP, Port: peerAddr.Port}
	c, err := peer.New(peerStruct, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peerAddr)
		return
	}
	defer c.Close()
	log.Printf("Completed handshake with %s\n", peerAddr)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)

	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workQueue)

	return buf, nil
}

func Open(path string) (*Torrent, error) {
	file, err := torrent.Open(path)
	if err != nil {
		return nil, err
	}

	peerID, err := generatePeerID()
	if err != nil {
		return nil, err
	}

	peers, err := requestPeers(file, peerID, Port)
	if err != nil {
		return nil, err
	}

	torrent := Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    file.InfoHash,
		PieceHashes: file.PieceHashes,
		PieceLength: file.PieceLength,
		Length:      file.Length,
		Name:        file.Name,
	}

	return &torrent, nil
}

func requestPeers(t *torrent.TorrentFile, peerID [20]byte, port uint16) ([]torrent.Peer, error) {
	resp, err := torrent.RequestPeers(t, peerID, port)
	if err != nil {
		return nil, err
	}

	return resp.Peers, nil
}

func generatePeerID() ([20]byte, error) {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	return peerID, err
}

func (t *Torrent) DownloadToFile(path string) error {
	var f *os.File
	var err error

	if path == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	buf, err := t.Download()
	if err != nil {
		return err
	}

	_, err = f.Write(buf)
	return err
}
