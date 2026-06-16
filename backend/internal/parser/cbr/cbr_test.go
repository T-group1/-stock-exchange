package cbr

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ИСПРАВЛЕНО: реалистичные тестовые данные
// Для JPY: Nominal=100, Value=55,6789 (курс за 100 иен), VunitRate=0,556789 (курс за 1 иену)
const testXML = `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="15.06.2026" name="Foreign Currency Market Rates">
	<Valute ID="R01010">
		<NumCode>036</NumCode>
		<CharCode>AUD</CharCode>
		<Nominal>1</Nominal>
		<Name>Австралийский доллар</Name>
		<Value>54,3210</Value>
		<VunitRate>54,3210</VunitRate>
	</Valute>
	<Valute ID="R01235">
		<NumCode>840</NumCode>
		<CharCode>USD</CharCode>
		<Nominal>1</Nominal>
		<Name>Доллар США</Name>
		<Value>82,5000</Value>
		<VunitRate>82,5000</VunitRate>
	</Valute>
	<Valute ID="R01239">
		<NumCode>978</NumCode>
		<CharCode>EUR</CharCode>
		<Nominal>1</Nominal>
		<Name>Евро</Name>
		<Value>90,1234</Value>
		<VunitRate>90,1234</VunitRate>
	</Valute>
	<Valute ID="R01820">
		<NumCode>392</NumCode>
		<CharCode>JPY</CharCode>
		<Nominal>100</Nominal>
		<Name>Японских иен</Name>
		<Value>55,6789</Value>
		<VunitRate>0,556789</VunitRate>
	</Valute>
</ValCurs>`

func TestParseValCurs(t *testing.T) {
	client := NewClient()

	rates, err := client.parseValCurs([]byte(testXML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем количество (4 валюты + RUB)
	if len(rates) != 5 {
		t.Errorf("expected 5 rates, got %d", len(rates))
	}

	// Проверяем парсинг даты
	if rates[0].Date != "2026-06-15" {
		t.Errorf("expected date 2026-06-15, got %s", rates[0].Date)
	}

	var usd, eur, jpy *ParsedRate
	for i := range rates {
		switch rates[i].CharCode {
		case "USD":
			usd = &rates[i]
		case "EUR":
			eur = &rates[i]
		case "JPY":
			jpy = &rates[i]
		}
	}

	if usd == nil {
		t.Fatal("USD not found")
	}
	if usd.Rate != 82.5 {
		t.Errorf("expected USD rate 82.5, got %f", usd.Rate)
	}
	if usd.Nominal != 1 {
		t.Errorf("expected USD nominal 1, got %d", usd.Nominal)
	}

	if eur == nil {
		t.Fatal("EUR not found")
	}
	if eur.Rate != 90.1234 {
		t.Errorf("expected EUR rate 90.1234, got %f", eur.Rate)
	}

	// ИСПРАВЛЕНО: JPY — VunitRate = 0.556789 (курс за 1 иену)
	if jpy == nil {
		t.Fatal("JPY not found")
	}
	if jpy.Nominal != 100 {
		t.Errorf("expected JPY nominal 100, got %d", jpy.Nominal)
	}
	if jpy.Rate != 0.556789 {
		t.Errorf("expected JPY rate 0.556789, got %f", jpy.Rate)
	}

	var rub *ParsedRate
	for i := range rates {
		if rates[i].CharCode == "RUB" {
			rub = &rates[i]
			break
		}
	}
	if rub == nil {
		t.Fatal("RUB not found")
	}
	if rub.Rate != 1.0 {
		t.Errorf("expected RUB rate 1.0, got %f", rub.Rate)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"82,5000", 82.5},
		{"90,1234", 90.1234},
		{"1,0000", 1.0},
		{" 0,556789 ", 0.556789},
	}

	for _, tt := range tests {
		got, err := parseFloat(tt.input)
		if err != nil {
			t.Errorf("parseFloat(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestGetDailyRates_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testXML))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/",
	}

	resp, err := client.httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to get from mock server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestParseCBRRate_WithNominal(t *testing.T) {
	valute := Valute{
		ID:        "R01700J",
		NumCode:   "392",
		CharCode:  "JPY",
		Nominal:   100,
		Name:      "Японских иен",
		Value:     "55,6789",
		VunitRate: "",
	}

	rate, err := parseCBRRate(valute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Value / Nominal = 55.6789 / 100 = 0.556789
	expected := 55.6789 / 100.0
	if rate.Rate != expected {
		t.Errorf("expected rate %f, got %f", expected, rate.Rate)
	}
}

func TestParseCBRRate_WithVunitRate(t *testing.T) {
	valute := Valute{
		ID:        "R01235",
		NumCode:   "840",
		CharCode:  "USD",
		Nominal:   1,
		Name:      "Доллар США",
		Value:     "82,5000",
		VunitRate: "82,5000",
	}

	rate, err := parseCBRRate(valute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rate.Rate != 82.5 {
		t.Errorf("expected rate 82.5, got %f", rate.Rate)
	}
}
