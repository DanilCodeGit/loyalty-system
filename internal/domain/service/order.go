package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// OrderRepository is an interface for orders repository.
type OrderRepository interface {
	AddOrderForUser(ctx context.Context, order entity.Order) error
	GetAllUserOrders(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error)
	UpdateOrderForUser(ctx context.Context, order entity.Order) error
	SaveUnprocessedOrder(ctx context.Context, order entity.Order) error
	GetUnprocessedOrders(ctx context.Context) ([]entity.Order, error)
}

type OrderClient interface {
	Check(ctx context.Context, number int) (entity.Order, string, error)
}

// OrderService is a service for managing orders.
type OrderService struct {
	repository OrderRepository
	logger     *logger.Logger
	client     OrderClient
	orderQueue chan entity.Order
}

// NewOrderService returns a new order service.
func NewOrderService(logger *logger.Logger, orderQueue chan entity.Order, repository OrderRepository) *OrderService {
	return &OrderService{
		repository: repository,
		logger:     logger,
		orderQueue: orderQueue,
	}
}

// SetClient sets the client for the order service.
func (s *OrderService) SetClient(client OrderClient) {
	s.client = client
}

// Add adds a new order for a user.
func (s *OrderService) Add(ctx context.Context, orderNumber int, userUUID uuid.UUID) error {
	const op = "domain.services.OrderService.Add"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	order := entity.NewOrder(userUUID, orderNumber)
	err := s.repository.AddOrderForUser(ctx, order)
	if err != nil {
		return err
	}

	s.orderQueue <- order
	log.Info("Order added to queue", log.AnyField("order_number", order.Number))
	return nil
}

// GetAll returns all orders for a user.
func (s *OrderService) GetAll(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error) {
	orders, err := s.repository.GetAllUserOrders(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// Check return count bonuses for the order
func (s *OrderService) Check(ctx context.Context, order entity.Order) (float64, error) {
	const op = "domain.services.OrderService.Check"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
		s.logger.AnyField("order_number", order.Number),
		s.logger.AnyField("user_uuid", order.UserUUID),
		s.logger.AnyField("uploaded_at", order.UploadedAt),
	)
	for {
		select {
		case <-ctx.Done():
			log.Info("Context is done")
			return 0, ctx.Err()
		default:
			externalOrder, respTimeOut, err := s.client.Check(ctx, order.Number)
			if err != nil {
				if errors.Is(err, entity.ErrExternalOrderRateLimitExceeded) {
					duration, err := time.ParseDuration(respTimeOut)
					if err != nil {
						return 0, err
					}
					time.Sleep(duration * time.Second)
					continue
				}
				if errors.Is(err, entity.ErrExternalOrderNotRegistered) {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				return 0, err
			}
			if externalOrder.Status == entity.OrderRegistered && order.Status != entity.OrderProcessing {
				log.Info("Order is processing")
				order.Status = entity.OrderProcessing
				err := s.Update(ctx, order)
				if err != nil {
					return 0, err
				}
				continue
			}
			if externalOrder.Status == entity.OrderProcessing && externalOrder.Status != order.Status {
				log.Info("Order is processing")
				order.Status = entity.OrderProcessing
				err := s.Update(ctx, order)
				if err != nil {
					return 0, err
				}
				continue
			}
			if externalOrder.Status == entity.OrderProcessed {
				log.Info("Order processed")
				order.Status = entity.OrderProcessed
				order.Accrual = externalOrder.Accrual
				err = s.Update(ctx, order)
				if err != nil {
					return 0, err
				}
				return order.Accrual, nil
			}
			if externalOrder.Status == entity.OrderInvalid {
				log.Info("Order invalid")
				order.Status = entity.OrderInvalid
				err = s.Update(ctx, order)
				if err != nil {
					return 0, err
				}
				return 0, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *OrderService) Update(ctx context.Context, order entity.Order) error {
	const op = "domain.services.OrderService.Update"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
		s.logger.AnyField("order_number", order.Number),
		s.logger.AnyField("user_uuid", order.UserUUID),
		s.logger.AnyField("uploaded_at", order.UploadedAt),
		s.logger.AnyField("status", order.Status),
	)

	err := s.repository.UpdateOrderForUser(ctx, order)
	if err != nil {
		return err
	}
	log.Info("Order updated")
	return nil
}

// Сохранение необработанного заказа в базу данных
func (s *OrderService) SaveUnprocessedOrder(order entity.Order) error {
	// Реализуйте логику для сохранения заказа в базу данных
	return s.repository.SaveUnprocessedOrder(context.Background(), order)
}

// Получение всех необработанных заказов из базы данных
func (s *OrderService) GetUnprocessedOrders() ([]entity.Order, error) {
	// Реализуйте логику для получения необработанных заказов из базы данных
	return s.repository.GetUnprocessedOrders(context.Background())
}
