package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/nek0ill/chago/internal/chat"
	"github.com/nek0ill/chago/internal/monitoring"
)

var (
	serverPort  string
	encryptionKey string
)

func main() {
	monitoring.InitMetrics()

	rootCmd := &cobra.Command{
		Use:   "chago",
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
		serverAddr  string
		key         string
	)

	clientCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a chat server",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := chat.NewClient(serverAddr)
			if err != nil {
				log.Fatal("Failed to connect:", err)
			}

			log.Printf("Connected to %s", serverAddr)
			
			// Start interactive session
			for {
				var msg string
				log.Print("Enter message: ")
				if _, err := fmt.Scanln(&msg); err != nil {
					log.Println("Error reading input:", err)
					continue
				}

				if err := client.SendMessage(msg, []byte(key)); err != nil {
					log.Println("Error sending message:", err)
				}
			}
		},
	}

	clientCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:8080", "Server address")
	clientCmd.Flags().StringVarP(&key, "key", "k", "", "Encryption key")
	clientCmd.Execute()
}
