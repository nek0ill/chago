package chat

import (
	"net"
	"time"
	"github.com/yourusername/encrypted-chat/internal/crypto"
	"github.com/yourusername/encrypted-chat/internal/monitoring"
)

type Server struct {
	listener    net.Listener
	clients     map[net.Conn]*Client
	encryptKey  []byte
}

func NewServer(port string) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		clients: make(map[net.Conn]*Client),
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
	defer conn.Close()
	monitoring.ActiveConnections.Inc()
	defer monitoring.ActiveConnections.Dec()

	client := &Client{conn: conn}
	s.clients[conn] = client

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			delete(s.clients, conn)
			return
		}

		decrypted, err := crypto.Decrypt(key, buf[:n])
		if err != nil {
			continue
		}

		monitoring.MessagesReceived.Inc()
		s.broadcast(decrypted, conn)
	}
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
