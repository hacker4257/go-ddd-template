package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/hacker4257/go-ddd-template/internal/api/http"
	"github.com/hacker4257/go-ddd-template/internal/api/http/handler"
	userapp "github.com/hacker4257/go-ddd-template/internal/app/user"
	"github.com/hacker4257/go-ddd-template/internal/infra/cache/redis"
	"github.com/hacker4257/go-ddd-template/internal/infra/mq/kafka"
	"github.com/hacker4257/go-ddd-template/internal/infra/persistence/mysql"
	"github.com/hacker4257/go-ddd-template/internal/pkg/config"
	"github.com/hacker4257/go-ddd-template/internal/pkg/health"
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
	)
	//mysql
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

	//redis
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

	//kafka
	kpub, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Error("kafka_producer_error", slog.Any("err", err))
		os.Exit(1)
	}
	defer kpub.Close()

	transactor := mysql.NewTransactor(db)
	outboxStore := mysql.NewOutboxStore(db)

	userCache := redis.NewUserCache(rdb)
	userRepo := mysql.NewUserRepo(db)
	userSvc := userapp.New(userRepo, userCache, cfg.Redis.UserTTL, transactor, outboxStore, cfg.Kafka.UserTopic)

	userHandler := handler.NewUserHandler(userSvc)

	readyHandler := handler.ReadyHandler{
	Checker: health.Checker{
		DB:      db,
		Redis:   rdb,
		Brokers: cfg.Kafka.Brokers,
		Timeout: 2 * time.Second,
	},
	}
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler: httpapi.NewRouter(log, userHandler, readyHandler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// 启动
	go func() {
		log.Info("http_server_start", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http_server_error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	// 优雅退出
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutdown_start")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown_error", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("shutdown_done")
}
