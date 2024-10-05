package top_n

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"tp1/internal/worker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
)

func New() (worker.Worker, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}
	_ = logs.InitLogger(cfg.String("log.level", "INFO"))
	b, err := amqpconn.NewBroker()
	if err != nil {
		return nil, err
	}

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	id, _ := strconv.Atoi(os.Getenv("worker-id"))

	return &Filter{
		id:         uint8(id),
		peers:      uint8(cfg.Int("exchange.peers", 1)),
		config:     cfg,
		broker:     b,
		signalChan: signalChan,
	}, nil
}
