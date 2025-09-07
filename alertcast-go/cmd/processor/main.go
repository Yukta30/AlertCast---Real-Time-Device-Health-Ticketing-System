package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"alertcast/internal/config"
	"alertcast/internal/kafka"
	"alertcast/internal/models"
	"alertcast/internal/repository"
	"alertcast/internal/cache"
	"alertcast/internal/rules"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reader := kafkaio.NewReader(cfg.KafkaBroker, cfg.ConsumerGroup, cfg.TopicEvents)
	defer kafkaio.CloseQuiet(reader)

	pg, err := repository.NewPostgres(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	rd := cache.New(cfg.RedisAddr)
	if err := rd.Ping(ctx); err != nil {
		log.Printf("redis ping: %v", err)
	}

	log.Printf("processor: broker=%s group=%s topic=%s", cfg.KafkaBroker, cfg.ConsumerGroup, cfg.TopicEvents)

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // shutdown
			}
			log.Printf("fetch: %v", err)
			continue
		}
		var ev models.DeviceEvent
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("unmarshal: %v", err)
			kafkaio.CommitQuiet(ctx, reader, m)
			continue
		}

		severity, reason, ok := rules.Evaluate(ev)
		if ok {
			b, _ := json.Marshal(ev)
			t := &models.Ticket{
				DeviceID:  ev.DeviceID,
				Severity:  severity,
				Reason:    reason,
				EventJSON: string(b),
			}
			if err := pg.InsertTicket(ctx, t); err != nil {
				log.Printf("insert: %v", err)
			} else {
				_ = rd.IncSeverity(ctx, severity)
				log.Printf("ticket #%d device=%s severity=%s reason=%s", t.ID, t.DeviceID, t.Severity, t.Reason)
			}
		}

		kafkaio.CommitQuiet(ctx, reader, m)
	}
}
