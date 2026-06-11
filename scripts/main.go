package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"T_Project/internal/db" // Путь к твоему сгенерированному пакету
	
	// Правильные импорты драйвера pgx v5
	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// 1. Строка подключения к твоей базе в Docker
	connStr := "postgres://rates_admin:rates_secure_pass@localhost:5432/rates_db?sslmode=disable"

	// 2. Устанавливаем соединение с СУБД (одиночное подключение)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer conn.Close(ctx)

	// 3. Инициализируем сгенерированный sqlc клиент
	queries := db.New(conn)

	// 4. ТЕСТ 1: Пробуем вставить новую валюту (например, Юань)
	err = queries.InsertCurrency(ctx, db.InsertCurrencyParams{
		Code: "CNY",
		Name: "Китайский юань",
	})
	if err != nil {
		log.Printf("Валюта уже есть или произошла ошибка: %v", err)
	}

	// 5. ТЕСТ 2: Имитируем работу Вовы Баркина (запись курса доллара на сегодня)
	testRate := db.SaveCurrencyRateParams{
		CurrencyCode: "USD",
		Rate:         92.4500, // Тестовый курс
		RateDate:     time.Now(),
	}

	err = queries.SaveCurrencyRate(ctx, testRate)
	if err != nil {
		log.Fatalf("Ошибка при вызове сгенерированной функции: %v", err)
	}

	fmt.Println("🎉 Успех! Go-код успешно подключился к Docker и записал курс через sqlc!")
}