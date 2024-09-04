package workers

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"loyalty/internal/domain/entity"
	"loyalty/internal/domain/service"
	httpc "loyalty/internal/infrastructure/http-client"
	"loyalty/internal/infrastructure/postgre"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

type OrderWorker struct {
	orderService *service.OrderService
	orderQueue   chan entity.Order
	logger       *logger.Logger
	errorChan    chan error
	ctx          context.Context
	balanceQueue chan entity.BalanceOperation
	db           *pgxpool.Pool
	accrualURL   string
}

func NewOrderWorker(ctx context.Context, logger *logger.Logger, orderQueue chan entity.Order, errorChanel chan error, balanceQueue chan entity.BalanceOperation, db *pgxpool.Pool, accrualURL string) *OrderWorker {
	return &OrderWorker{
		orderQueue:   orderQueue,
		balanceQueue: balanceQueue,
		logger:       logger,
		errorChan:    errorChanel,
		ctx:          ctx,
		db:           db,
		accrualURL:   accrualURL,
	}
}

func (w *OrderWorker) Run() {
	const op = "app.workers.OrderWorker.Run"
	log := w.logger.With(w.logger.StringField("op", op))

	client, err := httpc.NewOrderClient(w.accrualURL, w.logger)
	if err != nil {
		log.Error("Failed to create order client", log.ErrorField(err))
		os.Exit(1)
	}

	orderRepository := postgre.NewOrderRepository(w.db, w.logger)
	w.orderService = service.NewOrderService(w.logger, w.orderQueue, orderRepository)
	w.orderService.SetClient(client)

	// Загружаем необработанные заказы из базы данных
	unprocessedOrders, err := w.orderService.GetUnprocessedOrders()
	if err != nil {
		log.Info("unprocessed orders not found", log.ErrorField(err))
		//return
	}

	// Помещаем их в очередь для обработки
	for _, order := range unprocessedOrders {
		w.orderQueue <- order
	}

	for i := 1; i <= 3; i++ {
		go w.worker()
	}
	log.Info("Start 3 order workers")
}

// worker is a goroutine that is responsible for processing orders.
func (w *OrderWorker) worker() {
	const op = "app.workers.order"
	log := w.logger.With(w.logger.StringField("op", op))
	for {
		select {
		case <-w.ctx.Done():
			w.saveUnprocessedOrders()
			return
		case order, ok := <-w.orderQueue:
			if !ok {
				w.errorChan <- fmt.Errorf("order queue is closed")
				return
			}

			reqID := "req_order" + fmt.Sprintf("%d", order.Number)
			ctx := context.WithValue(w.ctx, contexter.RequestID, reqID)

			log.Info("Processing order", log.AnyField("order_number", order.Number), log)
			bonuses, err := w.orderService.Check(ctx, order)
			if err != nil {
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
				return
			}

			log.Info("Order processed",
				log.AnyField("order_number", order.Number),
				log.AnyField("bonuses", bonuses),
			)
			if bonuses > 0 {
				w.balanceQueue <- entity.NewBalanceOperation(order.UserUUID, bonuses, 0, order.Number)
			}
		}
	}
}

// saveUnprocessedOrders сохраняет необработанные заказы в базу данных
func (w *OrderWorker) saveUnprocessedOrders() {
	for {
		select {
		case order := <-w.orderQueue:
			// Сохраняем заказ в базу данных как необработанный
			err := w.orderService.SaveUnprocessedOrder(order)
			if err != nil {
				w.logger.Error("Failed to save unprocessed order", w.logger.ErrorField(err))
				continue
			}
		default:
			// Если нет больше заказов в очереди, выходим
			return
		}
	}
}
