package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"flag"
	"github.com/spf13/cobra"
	"github.com/yourusername/encrypted-chat/internal/chat"
	"github.com/yourusername/encrypted-chat/internal/crypto"
	"github.com/yourusername/encrypted-chat/internal/monitoring"
)

var (
	serverPort  string
	encryptionKey string
)

func main() {
	monitoring.InitMetrics()

	rootCmd := &cobra.Command{
		Use:   "encrypted-chat",
		Short: "End-to-end encrypted chat application",
	}

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run as chat server",
		Run: runServer,
	}

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Run as chat client",
		Run: runClient,
	}

	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "8080", "Server port")
	serverCmd.Flags().StringVarP(&encryptionKey, "key", "k", "", "Encryption key (leave empty for auto-generation)")

	rootCmd.AddCommand(serverCmd, clientCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	key := []byte(encryptionKey)
	if len(key) == 0 {
		key = []byte(monitoring.GenerateEncryptionKey(32))
		log.Printf("Using generated encryption key: %s\n", string(key))
	}

	server, err := chat.NewServer(serverPort)
	if err != nil {
		log.Fatal(err)
	}

	// Start metrics server
	go func() {
		http.Handle("/metrics", monitoring.MetricsHandler())
		log.Println("Metrics server started on :2112")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Server started on port", serverPort)
	go server.Start(key)

	<-stop
	log.Println("Shutting down server...")
}

func runClient(cmd *cobra.Command, args []string) {
	var (
		serverAddr string
		key        string
	)

	cmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:8080", "Server address")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Encryption key (must match server key)")
	cmd.ParseFlags(os.Args[2:])

	if key == "" {
		log.Fatal("Encryption key is required for client")
	}

	client, err := chat.NewClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", serverAddr, err)
	}

	log.Printf("Connected to %s", serverAddr)

	// Start message sender
	go func() {
		for {
			var msg string
			fmt.Print("> ")
			if _, err := fmt.Scanln(&msg); err != nil {
				log.Println("Error reading input:", err)
				continue
			}

			if err := client.SendMessage(msg, []byte(key)); err != nil {
				log.Println("Error sending message:", err)
			}
		}
	}()

	// Start message receiver
	for {
		msg, err := client.ReceiveMessage([]byte(key))
		if err != nil {
			log.Println("Disconnected from server:", err)
			return
		}
		fmt.Printf("\n< %s\n> ", msg)
	}
}
