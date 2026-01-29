package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	auditapp "github.com/hacker4257/go-ddd-template/internal/app/audit"
	"github.com/hacker4257/go-ddd-template/internal/infra/cache/redis"
	"github.com/hacker4257/go-ddd-template/internal/infra/idempotency"
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

	// ---------- MySQL ----------
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

	// ---------- Redis ----------
	rdb, err := redis.NewClient(redis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Error("redis_open_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer rdb.Close()

	// ---------- Kafka Producer ----------
	kpub, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Error("kafka_producer_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer kpub.Close()

	// ---------- Context + Shutdown ----------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	startWorkerHTTP(ctx, log, cfg.Worker.HTTP.Addr)

	// ---------- Outbox Dispatcher ----------
	outboxStore := mysql.NewOutboxStore(db)
	dispatcher := NewOutboxDispatcher(log, outboxStore, kpub)
	go dispatcher.Run(ctx)

	// ---------- Idempotency Store ----------
	idem := idempotency.New(rdb)

	// ---------- Audit Service ----------
	auditRepo := mysql.NewAuditRepo(db)
	auditSvc := auditapp.New(auditRepo)

	// ---------- Kafka Consumer ----------
	consumer, err := NewUserConsumer(
		log,
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup,
		cfg.Kafka.UserTopic,
		cfg.Kafka.UserDLQTopic,
		cfg.Kafka.MaxRetries,
		auditSvc,
		idem,
		kpub, // 用同一个 producer 做 requeue + DLQ
	)
	if err != nil {
		log.Error("kafka_consumer_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer consumer.Close()

	go consumer.Run(ctx)

	log.Info("worker_started")

	<-stop
	log.Info("worker_shutdown")
	cancel()
}
