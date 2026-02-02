package cache

import (
	"sync"

	"order-service/internal/models"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]*models.Order),
	}
}

func (c *Cache) Set(order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderID] = order
}

func (c *Cache) Get(orderID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[orderID]
	return order, exists
}

func (c *Cache) GetAll() []*models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]*models.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, order)
	}
	return orders
}

func (c *Cache) Delete(orderID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.orders, orderID)
}

func (c *Cache) LoadFromSlice(orders []models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range orders {
		c.orders[orders[i].OrderID] = &orders[i]
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = make(map[string]*models.Order)
}
