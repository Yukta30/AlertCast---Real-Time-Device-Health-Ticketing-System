# AlertCast – Real-Time Device Health Ticketing System (Go + Kafka)

A production-style, **upload-to-GitHub-ready** project that ingests real-time device events, applies
**severity rules**, stores tickets in **PostgreSQL**, keeps live counters in **Redis**, and exposes a small **HTTP API**.
Includes **Docker Compose** for one-command local spin-up, and **Kafdrop** to inspect Kafka topics.

> Tailored for a data/back-end engineer background (streaming, infra, reliability). Ships with rules and realistic sample data to hit **200+ events/day** by default (configurable).

---

## Features
- Go microservices: **ingestor**, **processor**, **api**
- Kafka pipeline (`device_events` → rules → tickets)
- **Severity rules** (critical/high/medium/low) based on status/temperature/etc.
- **PostgreSQL** persistence for tickets
- **Redis** counters for fast stats
- **Kafdrop** UI to browse topics/messages
- Clean, documented code; ready to extend

## Architecture
```
[Ingestor] ---> [Kafka: device_events] ---> [Processor] ---> [PostgreSQL: tickets]
                                        \-> [Redis: severity counters]
                                                              |
                                                            [API]
```
Kafdrop is available to inspect topics/partitions.

## Tech
Go, Kafka (segmentio/kafka-go), PostgreSQL (pgx), Redis (go-redis), Docker, Docker Compose.

## Quickstart
1. **Clone & configure**
   ```bash
   cp .env.example .env
   ```
   (Defaults are fine.)

2. **Run**
   ```bash
   docker compose up --build
   ```

3. **Endpoints & UIs**
   - API: http://localhost:8080
     - `GET /health`
     - `GET /tickets/recent?limit=50`
     - `GET /stats`
   - Kafdrop: http://localhost:9000

## Environment
See `.env.example`:
- `KAFKA_BROKER=kafka:9092`
- `TOPIC_EVENTS=device_events`
- `PG_*` for Postgres
- `REDIS_ADDR=redis:6379`
- `INGEST_RATE=0.2`  # events/second (≈ 12/min; tweak to match volume needs)

## Example
```bash
curl http://localhost:8080/health
curl http://localhost:8080/tickets/recent?limit=5
curl http://localhost:8080/stats
```

## Throughput / “Personalized” knobs
- Set `INGEST_RATE` (events per second) in `.env`.
- Tweak `internal/rules/rules.go` to increase/decrease severity share (e.g., more `offline` → more critical tickets).
- Processor is group-consumer; scale via `docker compose up --scale processor=3`.

## License
MIT
