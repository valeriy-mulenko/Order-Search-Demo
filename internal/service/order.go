package service

import (
	"context"
	"fmt"
	"log"

	"order-service/internal/cache"
	"order-service/internal/database"
	"order-service/internal/models"
)

type OrderService struct {
	db    *database.PostgresBase
	cache *cache.Cache
}

func NewOrderService(db *database.PostgresBase, cache *cache.Cache) *OrderService {
	return &OrderService{
		db:    db,
		cache: cache,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order *models.Order) error {
	// Сохраняем в БД
	if err := s.db.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order to DB: %w", err)
	}

	// Сохраняем в кэш
	s.cache.Set(order)

	log.Printf("Order %s saved successfully", order.OrderID)
	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	// Пытаемся получить из кэша
	if order, exists := s.cache.Get(orderID); exists {
		return order, nil
	}

	// Если нет в кэше, ищем в БД
	order, err := s.db.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order from DB: %w", err)
	}

	// Если нашли в БД, сохраняем в кэш
	if order != nil {
		s.cache.Set(order)
	}

	return order, nil
}

func (s *OrderService) LoadCacheFromDB(ctx context.Context) error {
	orders, err := s.db.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to load orders from DB: %w", err)
	}

	s.cache.LoadFromSlice(orders)
	log.Printf("Loaded %d orders into cache", len(orders))

	return nil
}
