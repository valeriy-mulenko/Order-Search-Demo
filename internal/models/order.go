package models

import (
	"time"
)

type Order struct {
	OrderID     string    `json:"order_id"`
	ClientID    int64     `json:"client_id"`
	Locale      string    `json:"locale"`
	Delivery    Delivery  `json:"delivery"`
	Payment     Payment   `json:"payment"`
	Items       []Product `json:"items"`
	DateCreated time.Time `json:"date_created"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Type    string `json:"type"`
	City    string `json:"city"`
	Address string `json:"address"`
}

type Payment struct {
	Transaction string  `json:"transaction_id"`
	Currency    string  `json:"currency"`
	Provider    string  `json:"provider"`
	Amount      float64 `json:"amount"`
	DatePay     int64   `json:"date_pay"`
	Bank        string  `json:"bank"`
}

type Product struct {
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Brand     string  `json:"brand"`
	Price     float64 `json:"price"`
	Size      string  `json:"size"`
	Quantity  int     `json:"quantity"`
}
