package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
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
	// Создание основного контекста с возможностью прерывания
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.NewLogger()

	log.Info("Loading configuration...")
	conf := config.MustLoadConfig()

	// Подключение к базе данных
	db, err := pgxpool.New(mainCtx, conf.DB)
	if err != nil {
		log.Error("Failed to connect to database", log.ErrorField(err))
		os.Exit(1)
	}

	// Создание каналов для обработки заказов и операций с балансом
	log.Info("Create order queue channel...")
	orderQueue := make(chan entity.Order, 100) // Канал для сервера
	defer close(orderQueue)

	log.Info("Create retryable order queue channel...")
	retryableOrderQueue := make(chan workers.RetryableOrder, 100) // Канал для воркеров
	defer close(retryableOrderQueue)

	log.Info("Create balance queue channel...")
	balanceQueue := make(chan entity.BalanceOperation, 100)
	defer close(balanceQueue)

	log.Info("Create error channel...")
	errorChan := make(chan error)
	defer close(errorChan)

	// Обработка ошибок в горутине
	go func() {
		for orderErr := range errorChan {
			log.Error("Error in worker", log.ErrorField(orderErr))
		}
		log.Info("Error channel is closed")
	}()

	// Создание и запуск HTTP-сервера
	log.Info("Creating HTTP server...")
	srv, err := server.New(mainCtx, conf, log, orderQueue, db) // Передаем orderQueue
	if err != nil {
		log.Error("Failed to create HTTP server", log.ErrorField(err))
		os.Exit(1)
	}
	srv.Run() // Запуск сервера

	// Горутин для перемещения заказов из orderQueue в retryableOrderQueue
	go func() {
		for order := range orderQueue {
			retryableOrderQueue <- workers.RetryableOrder{
				Order:   order,
				Retries: 0,
			}
		}
	}()

	// Создание WaitGroup для воркеров
	var wg sync.WaitGroup

	// Запуск воркера обработки заказов
	log.Info("Starting order worker...")
	orderWorker := workers.NewOrderWorker(mainCtx, log, retryableOrderQueue, errorChan, balanceQueue, db, conf.AccrualAdr)
	wg.Add(1)
	go func() {
		defer wg.Done()
		orderWorker.Run() // Выполнение работы воркера
	}()

	// Запуск воркера обработки балансов
	log.Info("Starting balance worker...")
	balanceWorker := workers.NewBalanceWorker(mainCtx, log, balanceQueue, errorChan, db)
	wg.Add(1)
	go func() {
		defer wg.Done()
		balanceWorker.Run() // Выполнение работы воркера
	}()

	// Ожидание сигнала завершения работы
	<-mainCtx.Done()
	log.Info("Received shutdown signal")

	// Ожидание завершения всех воркеров через WaitGroup
	wg.Wait()

	// Закрытие подключения к базе данных
	log.Info("Closing database connection...")
	db.Close()

	// Небольшая задержка для завершения всех операций
	time.Sleep(3 * time.Second)
	log.Info("Goodbye!")
}
