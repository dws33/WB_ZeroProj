package model

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          *Delivery `json:"delivery"`
	Payment           *Payment  `json:"payment"`
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
	if o == nil {
		return fmt.Errorf("order == nil (unset)")
	}
	return errors.Join(
		val.ValidateStruct(o,
			val.Field(&o.OrderUID, val.Required),
			val.Field(&o.TrackNumber, val.Required),
			val.Field(&o.Entry, val.Required),

			val.Field(&o.Delivery),
			val.Field(&o.Payment),
			val.Field(&o.Items, val.Length(1, 0)), // min len == 1, max len non-limited

			val.Field(&o.Locale, val.Required, validLocale),
			val.Field(&o.CustomerID, val.Required),
			val.Field(&o.DeliveryService, val.Required),
			val.Field(&o.ShardKey, val.Required),
			val.Field(&o.SmID, val.Required),
			val.Field(&o.DateCreated, val.Required),
			val.Field(&o.OofShard, val.Required),
		),
		goodTotalIsSumItemsTotalPrice(o),
	)
}

var validLocale = val.NewStringRuleWithError(govalidator.IsISO693Alpha2, is.ErrCountryCode2) // ozzo-validator/is.CountryCode2 use ISO3166 which not contains "en" (in low register)

func goodTotalIsSumItemsTotalPrice(o *Order) error {
	var itemsTotalPrice int
	for _, item := range o.Items {
		itemsTotalPrice += item.TotalPrice
	}
	if o.Payment.GoodsTotal != itemsTotalPrice {
		return errors.New("the total price does not match the amount of items (Order.Payment.GoodsTotal != itemsTotalPrice)")
	}
	return nil
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
	if d == nil {
		return fmt.Errorf("delivery == nil (unset)")
	}
	return val.ValidateStruct(d,
		val.Field(&d.Name, val.Required), //todo len?
		val.Field(&d.Phone, val.Required, is.E164),
		val.Field(&d.Zip, val.Required, val.Match(zipRegExp)),
		val.Field(&d.City, val.Required),    // todo some set?
		val.Field(&d.Address, val.Required), // todo some set?
		val.Field(&d.Region, val.Required),  // todo some set?
		val.Field(&d.Email, val.Required, is.Email),
	)
}

var zipRegExp = regexp.MustCompile("^[0-9]{7}$")

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

func (p *Payment) Validate() error {
	if p == nil {
		return fmt.Errorf("payment == nil (unset)")
	}
	return val.ValidateStruct(p,
		val.Field(&p.Transaction, val.Required), //todo must be same as order_uid?
		val.Field(&p.Currency, val.Required, is.CurrencyCode),
		val.Field(&p.Provider, val.Required),
		val.Field(&p.Amount, val.Required, val.Min(1), // todo or 0?

			// check amount consistent delivery cost and goods total
			val.By(func(_ any) error {
				if p.Amount != p.DeliveryCost+p.GoodsTotal {
					return fmt.Errorf("amount inconsistent delivery_cost and goods_total:  exept: %d, exist: %d",
						p.Amount, p.DeliveryCost+p.GoodsTotal)
				}
				return nil
			})),

		val.Field(&p.PaymentDT, val.Required), // todo or 0?
		val.Field(&p.Bank, val.Required),
		val.Field(&p.DeliveryCost, val.Required, val.Min(0)), // todo ?
		val.Field(&p.CustomFee, val.Min(0)),                  // todo ?
		val.Field(&p.GoodsTotal),
	)
}

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

func (i *Item) Validate() error {
	return val.ValidateStruct(i,
		val.Field(&i.ChrtID, val.Required),
		val.Field(&i.TrackNumber, val.Required),
		val.Field(&i.Price, val.Required, val.Min(0)), // todo min price > 0?
		val.Field(&i.RID, val.Required),
		val.Field(&i.Name, val.Required),
		val.Field(&i.Sale, val.Required, val.Min(0)),
		val.Field(&i.Size, val.Required), // todo?
		val.Field(&i.TotalPrice, val.Required, val.Min(0), // todo min price > 0?

			// check total price consistent price and sale
			val.By(func(_ any) error {
				discountedPrice := float64(i.Price) * (1 - float64(i.Sale)/100)
				mustPrice := int(math.Floor(discountedPrice))
				if mustPrice != i.TotalPrice {
					return fmt.Errorf("total price inconsistent price and sale: exept: %d, exist: %d",
						mustPrice, i.TotalPrice)
				}
				return nil
			})),

		val.Field(&i.NmID, val.Required),
		val.Field(&i.Brand, val.Required),
		val.Field(&i.Status, val.Required),
	)
}
