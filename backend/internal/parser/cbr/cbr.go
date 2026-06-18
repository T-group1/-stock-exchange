package cbr

import (
	"bytes"
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

// Браузерные заголовки, чтобы ЦБ не резал запрос как бота.
var browserHeaders = map[string]string{
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Accept":          "application/xml, text/xml, */*;q=0.8",
	"Accept-Language": "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
	"Accept-Encoding": "identity", // важно: не сжимаем, чтобы не возиться с gzip
	"Connection":      "keep-alive",
}

// CBRClient — интерфейс для работы с API ЦБ РФ
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
			Transport: &http.Transport{
				DisableCompression: true,
			},
		},
		baseURL: "https://www.cbr.ru/scripts/",
	}
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range browserHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBR returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// unmarshalCBRXML - умный парсер, который знает русскую кодировку
func unmarshalCBRXML(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unknown charset: %s", charset)
		}
	}
	return decoder.Decode(v)
}

// GetDailyRates получает курсы валют на указанную дату
func (c *Client) GetDailyRates(date time.Time) ([]ParsedRate, error) {
	url := c.baseURL + "XML_daily.asp"
	if !date.IsZero() {
		url = fmt.Sprintf("%s?date=%s", url, date.Format("02/01/2006"))
	}

	body, err := c.doRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily rates: %w", err)
	}

	return c.parseValCurs(body)
}

// GetAllCurrencies получает список всех доступных валют
func (c *Client) GetAllCurrencies() ([]ParsedRate, error) {
	url := c.baseURL + "XML_val.asp?d=0"

	body, err := c.doRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all currencies: %w", err)
	}

	return c.parseValuta(body)
}

func (c *Client) parseValuta(data []byte) ([]ParsedRate, error) {
	var valuta Valuta
	if err := unmarshalCBRXML(data, &valuta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Valuta XML: %w", err)
	}

	rates := make([]ParsedRate, 0, len(valuta.Items)+1)
	for _, item := range valuta.Items {
		rates = append(rates, ParsedRate{
			CharCode: item.CharCode,
			Nominal:  item.Nominal,
			Name:     item.Name,
			Rate:     0,
			Date:     "",
		})
	}

	rates = append(rates, ParsedRate{
		CharCode: "RUB",
		Nominal:  1,
		Name:     "Российский рубль",
		Rate:     0,
		Date:     "",
	})

	return rates, nil
}

func (c *Client) parseValCurs(data []byte) ([]ParsedRate, error) {
	var valCurs ValCurs
	if err := unmarshalCBRXML(data, &valCurs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	parsedDate, err := time.Parse(dateFormat, valCurs.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %q: %w", valCurs.Date, err)
	}
	isoDate := parsedDate.Format("2006-01-02")

	rates := make([]ParsedRate, 0, len(valCurs.Valutes)+1)
	for _, v := range valCurs.Valutes {
		rate, err := parseCBRRate(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse currency %s: %w", v.CharCode, err)
		}
		rate.Date = isoDate
		rates = append(rates, rate)
	}

	rates = append(rates, ParsedRate{
		CharCode: "RUB",
		Nominal:  1,
		Name:     "Российский рубль",
		Rate:     1.0,
		Date:     isoDate,
	})

	return rates, nil
}

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

func parseFloat(s string) (float64, error) {
	s = strings.Replace(s, ",", ".", 1)
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
