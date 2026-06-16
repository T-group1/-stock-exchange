package cbr

import "encoding/xml"

// ValCurs — корневой элемент XML от ЦБ РФ
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"` // Формат: DD.MM.YYYY
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

// Valute — одна валюта из XML
type Valute struct {
	ID        string `xml:"ID,attr"`   // Уникальный ID валюты (например, R01235)
	NumCode   string `xml:"NumCode"`   // Цифровой код (840 для USD)
	CharCode  string `xml:"CharCode"`  // Буквенный код (USD, EUR)
	Nominal   int    `xml:"Nominal"`   // Номинал (1, 10, 100)
	Name      string `xml:"Name"`      // Название на русском
	Value     string `xml:"Value"`     // Курс за номинал (строка, разделитель — запятая)
	VunitRate string `xml:"VunitRate"` // Курс за 1 единицу (строка, разделитель — запятая)
}

// ParsedRate — нормализованная структура курса после парсинга
type ParsedRate struct {
	CharCode string  // USD, EUR
	Nominal  int     // 1, 10, 100
	Name     string  // Название
	Rate     float64 // Курс за 1 единицу валюты
	Date     string  // YYYY-MM-DD (ISO формат)
}
