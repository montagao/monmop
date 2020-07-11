package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const apiURLv7 = `https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s`
const apiURLv7ExtraParams = `&range=1d&interval=5m&indicators=close&includeTimestamps=false&includePrePost=false&corsDomain=finance.yahoo.com&.tsrc=finance`

const noDataIndicator = `N/A`

// Ticker stores quote information for the particular ticker. The data
// for all the fields except 'Advancing' is fetched using Yahoo market API.
type Quote struct {
	Ticker     string  `json:"symbol"`                      // Stock ticker.
	LastTrade  float32 `json:"regularMarketPrice"`          // l1: last trade.
	Change     float32 `json:"regularMarketChange"`         // c6: change real time.
	ChangePct  float32 `json:"regularMarketChangePercent"`  // k2: percent change real time.
	Open       float32 `json:"regularMarketOpen"`           // o: market open price.
	Low        float32 `json:"regularMarketDayLow"`         // g: day's low.
	High       float32 `json:"regularMarketDayHigh"`        // h: day's high.
	Low52      float32 `json:"fiftyTwoWeekLow"`             // j: 52-weeks low.
	High52     float32 `json:"fiftyTwoWeekHigh"`            // k: 52-weeks high.
	Volume     float32 `json:"regularMarketVolume"`         // v: volume.
	AvgVolume  float32 `json:"averageDailyVolume10Day"`     // a2: average volume.
	PeRatio    float32 `json:"trailingPE"`                  // r2: P/E ration real time.
	PeRatioX   float32 `json:"trailingPE"`                  // r: P/E ration (fallback when real time is N/A).
	Dividend   float32 `json:"trailingAnnualDividendRate"`  // d: dividend.
	Yield      float32 `json:"trailingAnnualDividendYield"` // y: dividend yield.
	MarketCap  float32 `json:"marketCap"`                   // j3: market cap real time.
	MarketCapX float32 `json:"marketCap"`                   // j1: market cap (fallback when real time is N/A).
	Currency   string  `json:"currency"`                    // String code for currency of stock.
	PreOpen    float32 `json:"preMarketChangePercent,omitempty"`
	AfterHours float32 `json:"postMarketChangePercent,omitempty"`
}

// retrieve quotes for all tickers
func FetchAll(tickers []string) ([]Quote, error) {
	result := []Quote{}
	url := fmt.Sprintf(apiURLv7, strings.Join(tickers, `,`))

	response, err := http.Get(url + apiURLv7ExtraParams)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}

	result, err = unmarshalQuotes(body)
	if err != nil {
		return result, err
	}

	return result, nil
}

// retrieve quote for a single ticker
func FetchWithTicker(ticker string) (Quote, error) {
	result := Quote{}
	return result, nil
}

func unmarshalQuotes(body []byte) ([]Quote, error) {
	q := map[string]map[string][]Quote{}
	err := json.Unmarshal(body, &q)
	if err != nil {
		return nil, err
	}

	results := q["quoteResponse"]["result"]

	return results, nil
}
