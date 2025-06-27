package main

import (
	"context"
	"fmt"
	"github.com/dws33/WB_ZeroProj/internal/model"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"github.com/dws33/WB_ZeroProj/internal/storage/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
)

func main() {

	//connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	//	os.Getenv("PGHOST"),
	//	os.Getenv("PGPORT"),
	//	os.Getenv("PGUSER"),
	//	os.Getenv("PGPASSWORD"),
	//	os.Getenv("PGDATABASE"),
	//)
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "user", "password", "orders_db")

	ctx := context.TODO()

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgxPool.Close()
	err = pgxPool.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	dbStore := storage.NewStorage(pgxPool)
	cachedStore, err := cache.NewCachedStorage(ctx, dbStore)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/site/order", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.New("page").Parse(tpl))

		data := struct {
			Order *model.Order
			Error string
		}{}

		orderUID := r.URL.Query().Get("order_uid")
		order, err := cachedStore.GetOrder(ctx, orderUID)
		if err != nil {
			data.Error = fmt.Sprintf("Заказ с order_uid %q не найден", orderUID)
		} else {
			data.Order = order
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			log.Println("template execute error:", err)
		}
	})
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}

var tpl = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Поиск заказа по order_uid</title>
</head>
<body>
    <h1>Введите order_uid заказа</h1>
    <form method="GET" action="/site/order">
        <input type="text" name="order_uid" placeholder="order_uid" required>
        <input type="submit" value="Найти заказ">
    </form>

    {{if .Order}}
        <h2>Данные заказа:</h2>
        <p><b>Order UID:</b> {{.Order.OrderUID}}</p>
        <p><b>Track Number:</b> {{.Order.TrackNumber}}</p>
        <p><b>Entry:</b> {{.Order.Entry}}</p>
        <h3>Доставка</h3>
        <p>Имя: {{.Order.Delivery.Name}}</p>
        <p>Телефон: {{.Order.Delivery.Phone}}</p>
        <p>Город: {{.Order.Delivery.City}}</p>
        <p>Адрес: {{.Order.Delivery.Address}}</p>
        <p>Email: {{.Order.Delivery.Email}}</p>
        <h3>Оплата</h3>
        <p>Транзакция: {{.Order.Payment.Transaction}}</p>
        <p>Сумма: {{.Order.Payment.Amount}}</p>
        <p>Валюта: {{.Order.Payment.Currency}}</p>
        <h3>Товары</h3>
        <ul>
        {{range .Order.Items}}
            <li>{{.Name}} (Цена: {{.Price}}, Кол-во: 1, Общая цена: {{.TotalPrice}})</li>
        {{end}}
        </ul>
    {{else if .Error}}
        <p style="color:red;">{{.Error}}</p>
    {{end}}
</body>
</html>
`
