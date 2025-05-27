package chat

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
	"github.com/nek0ill/chago/internal/crypto"
	"github.com/nek0ill/chago/internal/monitoring"
)

const (
	maxMessageSize = 4096
)

type Server struct {
	listener    net.Listener
	clients     map[net.Conn]*Client
	encryptKey  []byte
	connectTime map[net.Conn]time.Time
	msgCount    map[net.Conn]int
}

func NewServer(port string) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener:    listener,
		clients:     make(map[net.Conn]*Client),
		connectTime: make(map[net.Conn]time.Time),
		msgCount:    make(map[net.Conn]int),
	}, nil
}

func (s *Server) Start(key []byte) {
	s.encryptKey = key
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn, s.encryptKey)
	}
}

func (s *Server) handleConnection(conn net.Conn, key []byte) {
	defer func() {
		conn.Close()
		delete(s.clients, conn)
		delete(s.connectTime, conn)
		log.Printf("Client disconnected: %s (%d messages)", conn.RemoteAddr(), s.msgCount[conn])
		delete(s.msgCount, conn)
	}()

	monitoring.ActiveConnections.Inc()
	defer monitoring.ActiveConnections.Dec()

	client := &Client{
		conn:       conn,
		encryptKey: key,
		decryptKey: key,
	}

	s.clients[conn] = client
	s.connectTime[conn] = time.Now()
	s.msgCount[conn] = 0
	log.Printf("New connection from: %s", conn.RemoteAddr())

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return
	}

	// Read message length prefix
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		log.Printf("Failed to read message length: %v", err)
		return
	}

	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > 4096 {
		log.Printf("Message too large: %d bytes", msgLen)
		return
	}

	// Read message payload  
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		log.Printf("Failed to read message: %v", err)
		return
	}

	decrypted, err := crypto.Decrypt(key, msgBuf)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		return
	}

	log.Printf("Received message (%d bytes)", len(msgBuf))
	monitoring.MessagesReceived.Inc()
	s.broadcast(decrypted, conn)
}

func (s *Server) broadcast(msg []byte, sender net.Conn) {
	encrypted, err := crypto.Encrypt(s.encryptKey, msg)
	if err != nil {
		log.Printf("Encryption failed during broadcast: %v", err)
		return
	}

	// Prepend message length (4 bytes)
	msgBuf := make([]byte, 4+len(encrypted))
	binary.BigEndian.PutUint32(msgBuf[0:4], uint32(len(encrypted)))
	copy(msgBuf[4:], encrypted)

	for conn := range s.clients {
		if conn != sender {
			err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				log.Printf("Failed to set write deadline: %v", err)
				delete(s.clients, conn)
				conn.Close()
				continue
			}

			if _, err := conn.Write(msgBuf); err != nil {
				log.Printf("Broadcast failed to %s: %v", conn.RemoteAddr(), err)
				delete(s.clients, conn)
				conn.Close()
			} else {
				log.Printf("Broadcasted message to %s", conn.RemoteAddr())
			}
		}
	}
}
