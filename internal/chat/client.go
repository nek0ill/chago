package chat

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"time"
	"github.com/nek0ill/chago/internal/crypto"
	"github.com/nek0ill/chago/internal/monitoring"
)

type Client struct {
	conn        net.Conn
	encryptKey  []byte
	decryptKey  []byte
	connected   bool
}

func NewClient(serverAddr string) (*Client, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) SendMessage(msg string, key []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("Failed to set write deadline: %v", err)
		return err
	}

	encrypted, err := crypto.Encrypt(key, []byte(msg))
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		return err
	}

	// Prepend message length (4 bytes)
	msgBuf := make([]byte, 4+len(encrypted))
	binary.BigEndian.PutUint32(msgBuf[0:4], uint32(len(encrypted)))
	copy(msgBuf[4:], encrypted)

	if _, err := c.conn.Write(msgBuf); err != nil {
		log.Printf("Failed to send message: %v", err)
		return err
	}

	monitoring.MessagesSent.Inc()
	log.Printf("Sent message (%d bytes)", len(msgBuf))
	return nil
}

func (c *Client) ReceiveMessage(key []byte) (string, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return "", err
	}

	// First read message length (4 bytes)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, lenBuf); err != nil {
		log.Printf("Failed to read message length: %v", err)
		return "", err
	}

	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > 4096 {
		log.Printf("Message too large: %d bytes", msgLen)
		return "", errors.New("message too large")
	}

	// Read message payload
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(c.conn, msgBuf); err != nil {
		log.Printf("Failed to read message: %v", err)
		return "", err
	}

	decrypted, err := crypto.Decrypt(key, msgBuf)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		return "", err
	}

	log.Printf("Received message (%d bytes)", len(msgBuf))
	return string(decrypted), nil
}
