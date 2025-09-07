package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"alertcast/internal/config"
	"alertcast/internal/kafka"
	"alertcast/internal/models"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	writer := kafkaio.NewWriter(cfg.KafkaBroker, cfg.TopicEvents)
	defer kafkaio.CloseQuiet(writer)

	log.Printf("ingestor: broker=%s topic=%s rate=%.3f eps", cfg.KafkaBroker, cfg.TopicEvents, cfg.IngestRate)

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	interval := time.Duration(float64(time.Second) / cfg.IngestRate)
	if cfg.IngestRate <= 0 {
		interval = 5 * time.Second
	}

	deviceIDs := make([]string, 200)
	for i := range deviceIDs {
		deviceIDs[i] = randomDeviceID(i + 1)
	}
	statuses := []string{"online", "online", "online", "degraded", "offline"} // bias towards online

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ev := models.DeviceEvent{
				DeviceID:    deviceIDs[rand.Intn(len(deviceIDs))],
				Status:      statuses[rand.Intn(len(statuses))],
				Temperature: 60 + rand.Float64()*35, // 60-95°C
				Region:      []string{"us-west", "us-east", "eu-central", "ap-south"}[rand.Intn(4)],
				Timestamp:   time.Now().UTC(),
			}
			if ev.Status == "online" && ev.Temperature < 80 && rand.Float64() < 0.85 {
				ev.Message = "Heartbeat OK"
			}

			b, _ := json.Marshal(ev)
			err := writer.WriteMessages(ctx, kafka.Message{Value: b})
			if err != nil {
				log.Printf("write: %v", err)
			}
		}
	}
}

func randomDeviceID(n int) string {
	return "dev-" + randomBase36(n+int(time.Now().UnixNano()%1000))
}

func randomBase36(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
