package server

import (
	"bytes"
	"testing"

	"github.com/benzjeremy/spotify-screensaver/audio"
)

func TestEncodeWSBinaryFrame(t *testing.T) {
	payload := []byte{10, 20, 30, 40, 50}
	frame := encodeWSBinaryFrame(payload)

	// Opcode 0x82 (FIN + binary)
	if frame[0] != 0x82 {
		t.Fatalf("Expected opcode 0x82, got 0x%02x", frame[0])
	}
	// Unmasked payload length 5
	if frame[1] != 5 {
		t.Fatalf("Expected length 5, got %d", frame[1])
	}
	if !bytes.Equal(frame[2:], payload) {
		t.Fatalf("Payload mismatch")
	}
}

func TestReadWSFrame(t *testing.T) {
	// Build a masked client text frame (FIN + Text 0x81, length 4, mask key 1 2 3 4)
	mask := []byte{1, 2, 3, 4}
	text := []byte("ping")
	maskedText := make([]byte, 4)
	for i := 0; i < 4; i++ {
		maskedText[i] = text[i] ^ mask[i]
	}

	buf := bytes.NewBuffer([]byte{0x81, 0x84})
	buf.Write(mask)
	buf.Write(maskedText)

	opcode, payload, err := readWSFrame(buf)
	if err != nil {
		t.Fatalf("readWSFrame failed: %v", err)
	}
	if opcode != 0x01 {
		t.Fatalf("Expected opcode 0x01, got 0x%02x", opcode)
	}
	if string(payload) != "ping" {
		t.Fatalf("Expected payload 'ping', got %s", string(payload))
	}
}

func TestAudioHubLifecycle(t *testing.T) {
	engine := audio.NewFallbackSynthesizer()
	engine.Start()
	defer engine.Stop()

	hub := NewAudioHub(engine)
	hub.Start()
	defer hub.Stop()

	bands := engine.GetBands()
	if len(bands) != 64 {
		t.Fatalf("Expected 64 bands")
	}
}
