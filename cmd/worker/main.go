package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hacker4257/go-ddd-template/internal/infra/mq/kafka"
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

	kpub, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Error("kafka_producer_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer kpub.Close()

	log.Info("worker_ready")


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 示例：每 5 秒打一条日志，证明 worker 常驻运行
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info("worker_tick")
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("worker_shutdown")
}
