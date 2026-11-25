package main

import (
	"log/slog"
	"os"
	"tekora/services/order/internal/adapters/psql"
	"tekora/services/order/internal/adapters/transport"
	"tekora/services/order/internal/config"
	"tekora/services/order/internal/core"
)

func main() {
	_ = config.MustConfig("services/user/.env")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("hello from order service")

	repo := psql.NewOrderRepository()
	svc := core.NewOrderService(repo)
	_ = transport.NewOrderServer(svc)

}
