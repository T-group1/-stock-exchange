package cbr

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	timeout    = 10 * time.Second
	dateFormat = "02.01.2006" // Формат даты ЦБ РФ
)

// CBRClient — интерфейс для работы с API ЦБ РФ (для тестируемости)
type CBRClient interface {
	GetDailyRates(date time.Time) ([]ParsedRate, error)
	GetAllCurrencies() ([]ParsedRate, error)
}

// Client — клиент для работы с API ЦБ РФ
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient создаёт новый клиент ЦБ РФ
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://www.cbr.ru/scripts/",
	}
}

// GetDailyRates получает курсы валют на указанную дату
func (c *Client) GetDailyRates(date time.Time) ([]ParsedRate, error) {
	url := c.baseURL + "XML_daily.asp"
	if !date.IsZero() {
		url = fmt.Sprintf("%s?date=%s", url, date.Format("02/01/2006"))
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBR returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// ЦБ РФ отдаёт XML в кодировке windows-1251, конвертируем в UTF-8
	utf8Body, err := decodeWindows1251(body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode windows-1251: %w", err)
	}

	return c.parseValCurs(utf8Body)
}

// GetAllCurrencies получает список всех доступных валют
func (c *Client) GetAllCurrencies() ([]ParsedRate, error) {
	url := c.baseURL + "XML_val.asp?d=0"

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all currencies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBR returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// ЦБ РФ отдаёт XML в кодировке windows-1251, конвертируем в UTF-8
	utf8Body, err := decodeWindows1251(body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode windows-1251: %w", err)
	}

	return c.parseValuta(utf8Body)
}

// decodeWindows1251 декодирует байты из кодировки windows-1251 в UTF-8
func decodeWindows1251(data []byte) ([]byte, error) {
	decoder := charmap.Windows1251.NewDecoder()
	utf8Bytes, err := decoder.Bytes(data)
	if err != nil {
		return nil, fmt.Errorf("windows-1251 decoding failed: %w", err)
	}
	return utf8Bytes, nil
}

// parseValuta парсит XML_val.asp (список валют без курсов)
func (c *Client) parseValuta(data []byte) ([]ParsedRate, error) {
	var valuta Valuta
	if err := xml.Unmarshal(data, &valuta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Valuta XML: %w", err)
	}

	rates := make([]ParsedRate, 0, len(valuta.Items)+1) // +1 для RUB
	for _, item := range valuta.Items {
		rates = append(rates, ParsedRate{
			CharCode: item.CharCode,
			Nominal:  item.Nominal,
			Name:     item.Name,
			Rate:     0,  // XML_val.asp не содержит курсов
			Date:     "", // XML_val.asp не содержит даты
		})
	}

	// ДОБАВЛЕНО: ЦБ РФ не включает RUB в XML_val.asp, добавляем вручную
	rates = append(rates, ParsedRate{
		CharCode: "RUB",
		Nominal:  1,
		Name:     "Российский рубль",
		Rate:     0,
		Date:     "",
	})

	return rates, nil
}

// parseValCurs парсит XML_daily.asp (курсы валют)
func (c *Client) parseValCurs(data []byte) ([]ParsedRate, error) {
	var valCurs ValCurs
	if err := xml.Unmarshal(data, &valCurs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	parsedDate, err := time.Parse(dateFormat, valCurs.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %q: %w", valCurs.Date, err)
	}
	isoDate := parsedDate.Format("2006-01-02")

	rates := make([]ParsedRate, 0, len(valCurs.Valutes)+1) // +1 для RUB
	for _, v := range valCurs.Valutes {
		rate, err := parseCBRRate(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse currency %s: %w", v.CharCode, err)
		}
		rate.Date = isoDate
		rates = append(rates, rate)
	}

	// Добавляем RUB с курсом 1.0
	rates = append(rates, ParsedRate{
		CharCode: "RUB",
		Nominal:  1,
		Name:     "Российский рубль",
		Rate:     1.0,
		Date:     isoDate,
	})

	return rates, nil
}

// parseCBRRate парсит одну валюту
func parseCBRRate(v Valute) (ParsedRate, error) {
	var rate float64
	var err error

	if v.VunitRate != "" {
		rate, err = parseFloat(v.VunitRate)
		if err != nil {
			return ParsedRate{}, fmt.Errorf("failed to parse VunitRate: %w", err)
		}
	} else {
		rate, err = parseFloat(v.Value)
		if err != nil {
			return ParsedRate{}, fmt.Errorf("failed to parse Value: %w", err)
		}
		if v.Nominal > 1 {
			rate = rate / float64(v.Nominal)
		}
	}

	return ParsedRate{
		CharCode: v.CharCode,
		Nominal:  v.Nominal,
		Name:     v.Name,
		Rate:     rate,
	}, nil
}

// parseFloat парсит число с запятой как разделителем (формат ЦБ РФ)
func parseFloat(s string) (float64, error) {
	s = strings.Replace(s, ",", ".", 1)
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
