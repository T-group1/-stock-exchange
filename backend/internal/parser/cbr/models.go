package cbr

import "encoding/xml"

// ValCurs — корневой элемент XML_daily.asp от ЦБ РФ (курсы валют)
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"` // Формат: DD.MM.YYYY
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

// Valuta — корневой элемент XML_val.asp от ЦБ РФ (список валют)
type Valuta struct {
	XMLName xml.Name     `xml:"Valuta"`
	Items   []ValutaItem `xml:"Item"`
}

// ValutaItem — одна валюта из XML_val.asp
type ValutaItem struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Name     string `xml:"Name"`
}

// Valute — одна валюта из XML_daily.asp
type Valute struct {
	ID        string `xml:"ID,attr"`
	NumCode   string `xml:"NumCode"`
	CharCode  string `xml:"CharCode"`
	Nominal   int    `xml:"Nominal"`
	Name      string `xml:"Name"`
	Value     string `xml:"Value"`
	VunitRate string `xml:"VunitRate"`
}

// ParsedRate — нормализованная структура курса после парсинга
type ParsedRate struct {
	CharCode string  // USD, EUR
	Nominal  int     // 1, 10, 100
	Name     string  // Название
	Rate     float64 // Курс за 1 единицу валюты
	Date     string  // YYYY-MM-DD (ISO формат)
}
