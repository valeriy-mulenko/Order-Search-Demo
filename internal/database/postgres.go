package database

import (
	"context"
	"fmt"

	"order-service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBase struct {
	pool *pgxpool.Pool
}

func NewPostgresBase(pool *pgxpool.Pool) *PostgresBase {
	return &PostgresBase{pool: pool}
}

func (r *PostgresBase) SaveOrder(ctx context.Context, order *models.Order) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Сохраняем основной заказ
	_, err = tx.Exec(ctx, `
        INSERT INTO orders (order_id, client_id, locale, date_created)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (order_id) DO UPDATE SET
            client_id = EXCLUDED.client_id,
            locale = EXCLUDED.locale,
            date_created = EXCLUDED.date_created
    `, order.OrderID, order.ClientID, order.Locale, order.DateCreated)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// 2. Сохраняем доставку
	_, err = tx.Exec(ctx, `
        INSERT INTO delivery (order_id, name, phone, email, type, city, address)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (order_id) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            email = EXCLUDED.email,
            type = EXCLUDED.type,
            city = EXCLUDED.city,
            address = EXCLUDED.address
    `, order.OrderID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Email, order.Delivery.Type,
		order.Delivery.City, order.Delivery.Address)

	if err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	// 3. Сохраняем платеж
	_, err = tx.Exec(ctx, `
        INSERT INTO payment (order_id, transaction_id, currency, provider, amount, date_pay, bank)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT DO NOTHING
    `, order.OrderID, order.Payment.Transaction, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.DatePay,
		order.Payment.Bank)

	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// 4. Сохраняем товары
	for _, item := range order.Items {
		// Сохраняем товар в таблицу item
		_, err = tx.Exec(ctx, `
            INSERT INTO item (product_id, name, brand, price, size)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (product_id) DO UPDATE SET
                name = EXCLUDED.name,
                brand = EXCLUDED.brand,
                price = EXCLUDED.price,
                size = EXCLUDED.size
        `, item.ProductID, item.Name, item.Brand, item.Price, item.Size)

		if err != nil {
			return fmt.Errorf("failed to save item: %w", err)
		}

		// Сохраняем связь заказа с товаром
		_, err = tx.Exec(ctx, `
            INSERT INTO items (order_id, product_id, quantity)
            VALUES ($1, $2, $3)
            ON CONFLICT DO NOTHING
        `, order.OrderID, item.ProductID, item.Quantity)

		if err != nil {
			return fmt.Errorf("failed to save order-item link: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresBase) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	// Получаем основной заказ
	var order models.Order
	err := r.pool.QueryRow(ctx, `
        SELECT order_id, client_id, locale, date_created
        FROM orders
        WHERE order_id = $1
    `, orderID).Scan(
		&order.OrderID, &order.ClientID, &order.Locale, &order.DateCreated,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Получаем доставку
	err = r.pool.QueryRow(ctx, `
        SELECT name, phone, email, type, city, address
        FROM delivery
        WHERE order_id = $1
    `, orderID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Email,
		&order.Delivery.Type, &order.Delivery.City, &order.Delivery.Address,
	)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	// Получаем платеж
	err = r.pool.QueryRow(ctx, `
        SELECT transaction_id, currency, provider, amount, date_pay, bank
        FROM payment
        WHERE order_id = $1
    `, orderID).Scan(
		&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.DatePay, &order.Payment.Bank,
	)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Получаем товары
	rows, err := r.pool.Query(ctx, `
        SELECT i.product_id, i.name, i.brand, i.price, i.size, it.quantity
        FROM items it
        JOIN item i ON it.product_id = i.product_id
        WHERE it.order_id = $1
    `, orderID)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	var items []models.Product
	for rows.Next() {
		var item models.Product
		err := rows.Scan(
			&item.ProductID, &item.Name, &item.Brand, &item.Price,
			&item.Size, &item.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil
}

func (r *PostgresBase) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	// Получаем все заказы (ограничиваем для производительности)
	rows, err := r.pool.Query(ctx, `
        SELECT order_id, client_id, locale, date_created
        FROM orders 
        ORDER BY date_created DESC 
        LIMIT 100
    `)

	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderID, &order.ClientID, &order.Locale, &order.DateCreated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Получаем доставку для каждого заказа
		err = r.pool.QueryRow(ctx, `
            SELECT name, phone, email, type, city, address
            FROM delivery
            WHERE order_id = $1
        `, order.OrderID).Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Email,
			&order.Delivery.Type, &order.Delivery.City, &order.Delivery.Address,
		)

		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to get delivery for order %s: %w", order.OrderID, err)
		}

		// Получаем платеж для каждого заказа
		err = r.pool.QueryRow(ctx, `
            SELECT transaction_id, currency, provider, amount, date_pay, bank
            FROM payment
            WHERE order_id = $1
        `, order.OrderID).Scan(
			&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.DatePay, &order.Payment.Bank,
		)

		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to get payment for order %s: %w", order.OrderID, err)
		}

		// Получаем товары для каждого заказа
		itemRows, err := r.pool.Query(ctx, `
            SELECT i.product_id, i.name, i.brand, i.price, i.size, it.quantity
            FROM items it
            JOIN item i ON it.product_id = i.product_id
            WHERE it.order_id = $1
        `, order.OrderID)

		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to get items for order %s: %w", order.OrderID, err)
		}

		var items []models.Product
		for itemRows.Next() {
			var item models.Product
			err := itemRows.Scan(
				&item.ProductID, &item.Name, &item.Brand, &item.Price,
				&item.Size, &item.Quantity,
			)
			if err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan item: %w", err)
			}
			items = append(items, item)
		}
		itemRows.Close()
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PostgresBase) InitDB(ctx context.Context) error {
	// Ваши таблицы уже созданы, проверяем их существование
	// (или просто пропускаем инициализацию, если таблицы уже есть)

	// Проверяем, существует ли таблица orders
	var exists bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'orders'
        )
    `).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		// Создаем таблицы если их нет
		_, err := r.pool.Exec(ctx, `
            CREATE TABLE orders (
                order_id VARCHAR(50) PRIMARY KEY,
                client_id INTEGER NOT NULL,
                locale VARCHAR(10),
                date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `)
		if err != nil {
			return fmt.Errorf("failed to create orders table: %w", err)
		}

		_, err = r.pool.Exec(ctx, `
            CREATE TABLE delivery (
                order_id VARCHAR(50) PRIMARY KEY REFERENCES orders(order_id) ON DELETE CASCADE,
                phone VARCHAR(20),
                email VARCHAR(255),
                type VARCHAR(10),
                city VARCHAR(100),
                address TEXT,
                name VARCHAR(255)
            )
        `)
		if err != nil {
			return fmt.Errorf("failed to create delivery table: %w", err)
		}

		_, err = r.pool.Exec(ctx, `
            CREATE TABLE item (
                product_id BIGINT PRIMARY KEY,
                size VARCHAR(255),
                price DECIMAL(10, 2),
                name VARCHAR(255),
                brand VARCHAR(255)
            )
        `)
		if err != nil {
			return fmt.Errorf("failed to create item table: %w", err)
		}

		_, err = r.pool.Exec(ctx, `
            CREATE TABLE items (
                items_id SERIAL PRIMARY KEY,
                order_id VARCHAR(50) NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
                product_id BIGINT NOT NULL REFERENCES item(product_id) ON DELETE RESTRICT,
                quantity INTEGER NOT NULL
            )
        `)
		if err != nil {
			return fmt.Errorf("failed to create items table: %w", err)
		}

		_, err = r.pool.Exec(ctx, `
            CREATE TABLE payment (
                payment_id BIGSERIAL PRIMARY KEY,
                order_id VARCHAR(50) NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
                transaction_id VARCHAR(50),
                currency VARCHAR(10) DEFAULT 'RUB',
                provider VARCHAR(50),
                amount DECIMAL(10, 2),
                date_pay BIGINT,
                bank VARCHAR(50)
            )
        `)
		if err != nil {
			return fmt.Errorf("failed to create payment table: %w", err)
		}

		// Создаем индексы
		_, err = r.pool.Exec(ctx, `
            CREATE INDEX idx_orders_client_id ON orders(client_id);
            CREATE INDEX idx_items_order_id ON items(order_id);
            CREATE INDEX idx_items_product_id ON items(product_id);
            CREATE INDEX idx_payment_order_id ON payment(order_id);
            CREATE INDEX idx_payment_transaction_id ON payment(transaction_id);
            CREATE INDEX idx_item_brand ON item(brand);
        `)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	return nil
}
