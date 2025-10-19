package service

import (
	"cache-app/internal/dto/response"
	"cache-app/internal/model"
	"cache-app/internal/repository"
	"context"
	"log"
	"sync"
	"time"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByUID(ctx context.Context, orderUID string) (*response.OrderResponse, error)
	RestoreCacheFromDB(ctx context.Context) error
}

type orderService struct {
	repo    repository.OrderRepository
	cache   map[string]*model.Order
	cacheMu sync.RWMutex
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	service := &orderService{
		repo:  repo,
		cache: make(map[string]*model.Order),
	}

	// Восстановление кеша при запуске
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := service.RestoreCacheFromDB(ctx); err != nil {
		log.Printf("Failed to restore cache: %v", err)
	}

	return service
}

func (s *orderService) CreateOrder(ctx context.Context, order *model.Order) error {
	// Сохраняем в БД
	if _, err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	// Сохраняем в кеш
	s.cacheMu.Lock()
	s.cache[order.OrderUID] = order
	s.cacheMu.Unlock()

	log.Printf("Order %s saved to cache and database", order.OrderUID)
	return nil
}

func (s *orderService) GetOrderByUID(ctx context.Context, orderUID string) (*response.OrderResponse, error) {
	// Пытаемся получить из кеша
	s.cacheMu.RLock()
	cachedOrder, exists := s.cache[orderUID]
	s.cacheMu.RUnlock()

	if exists {
		return s.modelToResponse(cachedOrder), nil
	}

	// Если нет в кеше, ищем в БД
	order, err := s.repo.GetOrderByUID(ctx, orderUID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кеш для будущих запросов
	s.cacheMu.Lock()
	s.cache[orderUID] = order
	s.cacheMu.Unlock()

	return s.modelToResponse(order), nil
}

func (s *orderService) RestoreCacheFromDB(ctx context.Context) error {
	orders, err := s.repo.GetAllOrders(ctx)
	if err != nil {
		return err
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	for _, order := range orders {
		s.cache[order.OrderUID] = order
	}

	log.Printf("Restored %d orders to cache", len(orders))
	return nil
}
func (s *orderService) modelToResponse(order *model.Order) *response.OrderResponse {
	return &response.OrderResponse{
		OrderUID:    order.OrderUID,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,
		Delivery: response.DeliveryResponse{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		},
		Payment: response.PaymentResponse{
			Transaction:  order.Payment.Transaction,
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       order.Payment.Amount,
			PaymentDT:    order.Payment.PaymentDT,
			Bank:         order.Payment.Bank,
			DeliveryCost: order.Payment.DeliveryCost,
			GoodsTotal:   order.Payment.GoodsTotal,
		},
		Items:       s.itemsToResponse(order.Items),
		DateCreated: order.DateCreated,
		OofShard:    order.OofShard,
	}
}

func (s *orderService) itemsToResponse(items []model.Item) []response.ItemResponse {
	var responseItems []response.ItemResponse
	for _, item := range items {
		responseItems = append(responseItems, response.ItemResponse{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}
	return responseItems
}
