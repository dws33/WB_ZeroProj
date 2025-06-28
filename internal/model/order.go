package model

import (
	"errors"
	"time"

	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []*Item   `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	OofShard          string    `json:"oof_shard"`
	DateCreated       time.Time `json:"date_created"` // можно заменить на time.Time
}

func (o *Order) Validate() error {
	return errors.Join(
		val.ValidateStruct(o,
			val.Field(&o.OrderUID, val.Required),
			val.Field(&o.TrackNumber, val.Required),
			val.Field(&o.Entry, val.Required),

			val.Field(&o.Locale, val.Required, is.CountryCode2), //todo CountryCode2?
			val.Field(&o.CustomerID, val.Required),
			val.Field(&o.DeliveryService, val.Required),
			val.Field(&o.ShardKey, val.Required),
			val.Field(&o.SmID, val.Required),
			val.Field(&o.DateCreated, val.Required), // todo not future
			val.Field(&o.OofShard, val.Required),
		),
		o.Delivery.Validate(),
	)
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

func (d *Delivery) Validate() error {
	return val.ValidateStruct(d,
		val.Field(&d.Name, val.Required), //todo len?
		val.Field(&d.Phone, val.Required, is.E164),
		val.Field(&d.Zip, val.Required, is.UTFDigit), // todo len?
		val.Field(&d.City, val.Required),             // todo some set?
		val.Field(&d.Address, val.Required),          // todo some set?
		val.Field(&d.Region, val.Required),           // todo some set?
		val.Field(&d.Email, val.Required, is.Email),  // todo some set?
	)
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// todo calculate items total_size
func (p *Payment) Validate() error {
	return val.ValidateStruct(p,
		val.Field(&p.Transaction, val.Required), //todo same as order_uid
		val.Field(&p.Currency, val.Required, is.CurrencyCode),
		val.Field(&p.Provider, val.Required),
		val.Field(&p.Amount, val.Required, val.Min(1)), // todo or 0?
		val.Field(&p.PaymentDT, val.Required),          // todo or 0?
		val.Field(&p.Bank, val.Required),
		val.Field(&p.DeliveryCost, val.Required, val.Min(0)),
		val.Field(&p.CustomFee, val.Min(0)),
		val.Field(&p.GoodsTotal), // todo can order contain less then 1 items?
	)
}

// todo calculate total_price
type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
