package peer

import (
	"bytes"
	"testing"
)

func TestHandshakeSerialize(t *testing.T) {
	infoHash := [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}
	peerID := [20]byte{45, 68, 69, 49, 51, 52, 48, 45, 106, 80, 199, 219, 129, 99, 14, 116, 226, 131, 207, 249}

	h := NewHandshake(infoHash, peerID)
	buf := h.Serialize()

	expected := []byte{19} // length of "BitTorrent protocol"
	expected = append(expected, []byte("BitTorrent protocol")...)
	expected = append(expected, make([]byte, 8)...) // reserved
	expected = append(expected, infoHash[:]...)
	expected = append(expected, peerID[:]...)

	if !bytes.Equal(buf, expected) {
		t.Errorf("Handshake serialization mismatch")
	}
}

func TestHandshakeRoundTrip(t *testing.T) {
	infoHash := [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}
	peerID := [20]byte{45, 68, 69, 49, 51, 52, 48, 45, 106, 80, 199, 219, 129, 99, 14, 116, 226, 131, 207, 249}

	h1 := NewHandshake(infoHash, peerID)
	buf := h1.Serialize()

	h2, err := ReadHandshake(bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("Failed to read handshake: %v", err)
	}

	if h1.Pstr != h2.Pstr {
		t.Errorf("Pstr mismatch: %s != %s", h1.Pstr, h2.Pstr)
	}

	if h1.InfoHash != h2.InfoHash {
		t.Errorf("InfoHash mismatch")
	}

	if h1.PeerID != h2.PeerID {
		t.Errorf("PeerID mismatch")
	}
}

func TestMessageSerialize(t *testing.T) {
	tests := []struct {
		message  *Message
		expected []byte
	}{
		// Keep-alive message
		{nil, []byte{0, 0, 0, 0}},
		// Choke message (length=1, id=0)
		{&Message{ID: MsgChoke, Payload: []byte{}}, []byte{0, 0, 0, 1, 0}},
		// Have message (length=5, id=4, payload=piece index 4)
		{NewHaveMessage(4), []byte{0, 0, 0, 5, 4, 0, 0, 0, 4}},
	}

	for _, test := range tests {
		result := test.message.Serialize()
		if !bytes.Equal(result, test.expected) {
			t.Errorf("Message serialization mismatch.\nExpected: %v\nGot: %v", test.expected, result)
		}
	}
}

func TestMessageRoundTrip(t *testing.T) {
	tests := []*Message{
		nil, // keep-alive
		{ID: MsgChoke, Payload: []byte{}},
		{ID: MsgUnchoke, Payload: []byte{}},
		{ID: MsgInterested, Payload: []byte{}},
		{ID: MsgNotInterested, Payload: []byte{}},
		NewHaveMessage(42),
		NewRequestMessage(1, 0, 16384),
	}

	for _, original := range tests {
		buf := original.Serialize()
		parsed, err := ReadMessage(bytes.NewReader(buf))
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		if original == nil && parsed == nil {
			continue // both nil (keep-alive), test passed
		}

		if original == nil || parsed == nil {
			t.Errorf("One message is nil but the other is not. Original: %v, Parsed: %v", original, parsed)
			continue
		}

		if original.ID != parsed.ID {
			t.Errorf("Message ID mismatch: %d != %d", original.ID, parsed.ID)
		}

		if !bytes.Equal(original.Payload, parsed.Payload) {
			t.Errorf("Payload mismatch: %v != %v", original.Payload, parsed.Payload)
		}
	}
}

func TestBitfield(t *testing.T) {
	bf := Bitfield{0b01010100, 0b01010100}

	expected := []int{1, 3, 5, 9, 11, 13}
	unexpected := []int{0, 2, 4, 6, 7, 8, 10, 12, 14, 15}

	for _, i := range expected {
		if !bf.HasPiece(i) {
			t.Errorf("Expected bitfield to have piece %d", i)
		}
	}

	for _, i := range unexpected {
		if bf.HasPiece(i) {
			t.Errorf("Expected bitfield to not have piece %d", i)
		}
	}
}

func TestBitfieldSetPiece(t *testing.T) {
	bf := Bitfield{0b00000000, 0b00000000}

	bf.SetPiece(1)
	bf.SetPiece(9)

	expected := Bitfield{0b01000000, 0b01000000}

	if !bytes.Equal(bf, expected) {
		t.Errorf("Bitfield mismatch after setting pieces. Expected: %08b %08b, Got: %08b %08b",
			expected[0], expected[1], bf[0], bf[1])
	}
}

func TestParsePieceMessage(t *testing.T) {
	index := 4
	begin := 567
	block := []byte("hello world")

	msg := NewPieceMessage(index, begin, block)
	parsedBegin, parsedBlock, err := ParsePieceMessage(index, msg.Payload)

	if err != nil {
		t.Fatalf("ParsePieceMessage failed: %v", err)
	}

	if parsedBegin != begin {
		t.Errorf("Begin mismatch: expected %d, got %d", begin, parsedBegin)
	}

	if !bytes.Equal(parsedBlock, block) {
		t.Errorf("Block mismatch: expected %s, got %s", string(block), string(parsedBlock))
	}
}

func TestParseHaveMessage(t *testing.T) {
	expectedIndex := 42
	msg := NewHaveMessage(expectedIndex)

	parsedIndex, err := ParseHaveMessage(msg.Payload)
	if err != nil {
		t.Fatalf("ParseHaveMessage failed: %v", err)
	}

	if parsedIndex != expectedIndex {
		t.Errorf("Index mismatch: expected %d, got %d", expectedIndex, parsedIndex)
	}
}

func TestMessageName(t *testing.T) {
	testCases := []struct {
		message  *Message
		expected string
	}{
		{nil, "KeepAlive"},
		{&Message{ID: MsgChoke}, "Choke"},
		{&Message{ID: MsgUnchoke}, "Unchoke"},
		{&Message{ID: MsgInterested}, "Interested"},
		{&Message{ID: MsgNotInterested}, "NotInterested"},
		{&Message{ID: MsgHave}, "Have"},
		{&Message{ID: MsgBitfield}, "Bitfield"},
		{&Message{ID: MsgRequest}, "Request"},
		{&Message{ID: MsgPiece}, "Piece"},
		{&Message{ID: MsgCancel}, "Cancel"},
		{&Message{ID: MessageID(99)}, "Unknown#99"},
	}

	for _, tc := range testCases {
		result := tc.message.Name()
		if result != tc.expected {
			t.Errorf("Expected message name %s, got %s", tc.expected, result)
		}
	}
}
