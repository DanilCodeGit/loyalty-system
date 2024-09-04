package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"loyalty/config"
	"loyalty/internal/app/http-server/server"
	"loyalty/internal/app/workers"
	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/logger"
)

// run swag init -g internal/app/http-server/server/server.go to generate swagger docs
// run swag fmt -g internal/app/http-server/server/server.go to format swagger docs
func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.NewLogger()

	log.Info("Loading configuration...")
	conf := config.MustLoadConfig()

	db, err := pgxpool.New(mainCtx, conf.DB)
	if err != nil {
		log.Error("Failed to connect to database", log.ErrorField(err))
		os.Exit(1)
	}
	go func() {
		<-mainCtx.Done()
		log.Info("Closing database connection...")
		db.Close()
	}()

	log.Info("Create order queue chanel...")
	orderQueue := make(chan entity.Order, 100)
	defer close(orderQueue)

	log.Info("Create balance queue chanel...")
	balanceQueue := make(chan entity.BalanceOperation, 100)
	defer close(balanceQueue)

	log.Info("Create error chanel ...")
	errorChan := make(chan error)
	defer close(errorChan)
	go func() {
		for orderErr := range errorChan {
			log.Error("Error in worker", log.ErrorField(orderErr))
			//os.Exit(1)
		}
		log.Info("Error chanel is closed")
	}()

	log.Info("Creating HTTP server...")
	srv, err := server.New(mainCtx, conf, log, orderQueue, db)
	if err != nil {
		log.Error("Failed to create HTTP server", log.ErrorField(err))
		os.Exit(1)
	}
	srv.Run()

	orderWorker := workers.NewOrderWorker(mainCtx, log, orderQueue, errorChan, balanceQueue, db, conf.AccrualAdr)
	orderWorker.Run()

	balanceWorker := workers.NewBalanceWorker(mainCtx, log, balanceQueue, errorChan, db)
	balanceWorker.Run()

	<-mainCtx.Done()
	time.Sleep(3 * time.Second)
	log.Info("Good bye!")
}
