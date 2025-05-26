package monitoring

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once
)

var (
	MessagesSent = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "chat_messages_sent_total",
			Help: "Total number of messages sent",
		},
	)

	MessagesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "chat_messages_received_total",
			Help: "Total number of messages received",
		},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "chat_active_connections",
			Help: "Current number of active connections",
		},
	)
)

func GenerateEncryptionKey(length int) string {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}

func InitMetrics() {
	prometheus.MustRegister(MessagesSent)
	prometheus.MustRegister(MessagesReceived)
	prometheus.MustRegister(ActiveConnections)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
