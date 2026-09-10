package server

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/benzjeremy/spotify-screensaver/audio"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type wsClient struct {
	conn net.Conn
	rw   *bufio.ReadWriter
	send chan []byte
}

// AudioHub manages active WebSocket connections and broadcasts 60 FPS spectrum frames.
type AudioHub struct {
	mu          sync.RWMutex
	clients     map[*wsClient]bool
	audioEngine audio.Engine
	stopChan    chan struct{}
}

func NewAudioHub(engine audio.Engine) *AudioHub {
	return &AudioHub{
		clients:     make(map[*wsClient]bool),
		audioEngine: engine,
		stopChan:    make(chan struct{}),
	}
}

func (h *AudioHub) Start() {
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()

		for {
			select {
			case <-h.stopChan:
				return
			case <-ticker.C:
				bands := h.audioEngine.GetBands()
				data := make([]byte, audio.NumBands)
				copy(data, bands[:])

				h.broadcast(data)
			}
		}
	}()
}

func (h *AudioHub) Stop() {
	select {
	case <-h.stopChan:
	default:
		close(h.stopChan)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.conn.Close()
		delete(h.clients, c)
	}
}

func (h *AudioHub) register(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *AudioHub) unregister(c *wsClient) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}

func (h *AudioHub) broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Client buffer full, skip frame to maintain 60 FPS fluidity
		}
	}
}

// HandleAudioStream handles RFC 6455 WebSocket handshakes and client loops.
func (s *Server) HandleAudioStream(w http.ResponseWriter, r *http.Request) {
	// 1. Anti-CSWSH: Verify Origin
	origin := r.Header.Get("Origin")
	if origin != "" {
		if !strings.HasPrefix(origin, "http://127.0.0.1:") && !strings.HasPrefix(origin, "http://localhost:") {
			http.Error(w, "Forbidden: Invalid origin", http.StatusForbidden)
			return
		}
	}

	// 2. Token Authentication
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("X-Session-Token")
	}
	if token != s.sessionToken {
		http.Error(w, "Unauthorized: Invalid session token", http.StatusUnauthorized)
		return
	}

	// 3. RFC 6455 Upgrade Verification
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" ||
		!strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		http.Error(w, "Expected WebSocket Upgrade", http.StatusBadRequest)
		return
	}

	secKey := r.Header.Get("Sec-WebSocket-Key")
	if secKey == "" {
		http.Error(w, "Missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	// Calculate Sec-WebSocket-Accept
	h := sha1.New()
	h.Write([]byte(secKey + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 4. Hijack TCP connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Server does not support hijacking", http.StatusInternalServerError)
		return
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send HTTP 101 Switching Protocols response
	resp := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)

	if _, err := rw.WriteString(resp); err != nil {
		conn.Close()
		return
	}
	if err := rw.Flush(); err != nil {
		conn.Close()
		return
	}

	client := &wsClient{
		conn: conn,
		rw:   rw,
		send: make(chan []byte, 16),
	}

	s.wsHub.register(client)
	defer func() {
		s.wsHub.unregister(client)
		conn.Close()
	}()

	// Write loop: sends binary frames (Opcode 0x82)
	go func() {
		for payload := range client.send {
			frame := encodeWSBinaryFrame(payload)
			client.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if _, err := rw.Write(frame); err != nil {
				client.conn.Close()
				return
			}
			if err := rw.Flush(); err != nil {
				client.conn.Close()
				return
			}
		}
	}()

	// Read loop: handles close and ping/pong frames
	for {
		client.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		opcode, _, err := readWSFrame(rw.Reader)
		if err != nil {
			break
		}
		if opcode == 0x08 { // Close frame
			break
		}
		if opcode == 0x09 { // Ping frame -> Respond with Pong (0x0A)
			pong := []byte{0x8A, 0x00}
			rw.Write(pong)
			rw.Flush()
		}
	}
}

func encodeWSBinaryFrame(payload []byte) []byte {
	length := len(payload)
	var header []byte

	// 0x82 = FIN bit (0x80) + Binary Frame (0x02)
	if length < 126 {
		header = []byte{0x82, byte(length)}
	} else if length <= 65535 {
		header = []byte{0x82, 126, byte(length >> 8), byte(length & 0xFF)}
	} else {
		header = []byte{0x82, 127, 0, 0, 0, 0,
			byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length & 0xFF)}
	}

	return append(header, payload...)
}

func readWSFrame(r io.Reader) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}

	opcode := header[0] & 0x0F
	masked := (header[1] & 0x80) != 0
	length := int(header[1] & 0x7F)

	if length == 126 {
		lenBuf := make([]byte, 2)
		if _, err := io.ReadFull(r, lenBuf); err != nil {
			return 0, nil, err
		}
		length = int(lenBuf[0])<<8 | int(lenBuf[1])
	} else if length == 127 {
		lenBuf := make([]byte, 8)
		if _, err := io.ReadFull(r, lenBuf); err != nil {
			return 0, nil, err
		}
		length = int(lenBuf[4])<<24 | int(lenBuf[5])<<16 | int(lenBuf[6])<<8 | int(lenBuf[7])
	}

	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(r, mask[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := 0; i < length; i++ {
			payload[i] ^= mask[i%4]
		}
	}

	return opcode, payload, nil
}
