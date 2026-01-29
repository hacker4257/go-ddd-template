package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/hacker4257/go-ddd-template/internal/infra/mq/kafka"
	"github.com/hacker4257/go-ddd-template/internal/infra/persistence/mysql"
	"github.com/hacker4257/go-ddd-template/internal/pkg/config"
	"github.com/hacker4257/go-ddd-template/internal/pkg/logger"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.Log.Level).With(
		slog.String("app", cfg.App.Name),
		slog.String("env", cfg.App.Env),
		slog.String("proc", "worker"),
	)

	// MySQL
	db, err := mysql.Open(mysql.Config{
		DSN:             cfg.DB.MySQL.DSN,
		MaxOpenConns:    cfg.DB.MySQL.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MySQL.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.MySQL.ConnMaxLifetime,
	})
	if err != nil {
		log.Error("db_open_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()

	// Kafka Producer
	kpub, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Error("kafka_producer_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer kpub.Close()

	// Outbox store + dispatcher
	store := mysql.NewOutboxStore(db)
	dispatcher := NewOutboxDispatcher(log, store, kpub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go dispatcher.Run(ctx)
	log.Info("worker_started")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("worker_shutdown")
	cancel()
}
