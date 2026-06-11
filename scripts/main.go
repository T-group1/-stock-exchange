package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"T_Project/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://rates_admin:rates_secure_pass@localhost:5432/rates_db?sslmode=disable"

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	// ТЕСТ 1: Вставка валюты
	err = queries.InsertCurrency(ctx, db.InsertCurrencyParams{
		Code: "CNY",
		Name: "Китайский юань",
	})
	if err != nil {
		log.Printf("Валюта уже есть или произошла ошибка: %v", err)
	}

	// ТЕСТ 2: Запись курса доллара с правильными типами pgtype
	testRate := db.SaveCurrencyRateParams{
		CurrencyCode: pgtype.Text{
			String: "USD",
			Valid:  true,
		},
		Rate: pgtype.Numeric{
			Int:   big.NewInt(924500),
			Exp:   -4,
			Valid: true,
		},
		RateDate: time.Now(),
	}

	err = queries.SaveCurrencyRate(ctx, testRate)
	if err != nil {
		log.Fatalf("Ошибка при вызове сгенерированной функции: %v", err)
	}

	fmt.Println("🎉 Успех! Go-код успешно подключился к Docker и записал курс через sqlc!")
}