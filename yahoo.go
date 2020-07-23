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

var marketTickers = []string{
	"^DJI",     // Dow Jones
	"^GSPC",    // S&P 500
	"^IXIC",    // NASDAQ
	"^N225",    // Nikkei
	"^HSI",     // Hong Kong
	"^FTSE",    // London
	"^GDAXI",   // Frankfurt
	"^TNX",     // 10-Year Yield
	"CAD=X",    // CAD
	"EURUSD=X", // Euro
	"RMB=F",    // RMB
	"CL=F",     // Oil
	"GC=F",     // Gold
	"BTC-USD",  // Bitcoin
}

// Ticker stores quote information for the particular ticker. The data
// for all the fields except 'Advancing' is fetched using Yahoo market API.
type Quote struct {
	Ticker    string  `json:"symbol"`                     // Stock ticker.
	LastTrade float64 `json:"regularMarketPrice"`         // l1: last trade.
	Change    float64 `json:"regularMarketChange"`        // c6: change real time.
	ChangePct float64 `json:"regularMarketChangePercent"` // k2: percent change real time.
	Open      float64 `json:"regularMarketOpen"`          // o: market open price.
	Low       float64 `json:"regularMarketDayLow"`        // g: day's low.
	High      float64 `json:"regularMarketDayHigh"`       // h: day's high.
	// Low52      float64 `json:"fiftyTwoWeekLow"`             // j: 52-weeks low.
	// High52     float64 `json:"fiftyTwoWeekHigh"`            // k: 52-weeks high.
	Volume    float64 `json:"regularMarketVolume"`     // v: volume.
	AvgVolume float64 `json:"averageDailyVolume10Day"` // a2: average volume.
	PeRatio   float64 `json:"trailingPE"`              // r2: P/E ration real time.
	// PeRatioX   float64 `json:"trailingPE"`                  // r: P/E ration (fallback when real time is N/A).
	Dividend float64 `json:"trailingAnnualDividendYield"` // d: dividend.
	// Yield      float64 `json:"trailingAnnualDividendYield"` // y: dividend yield.
	MarketCap float64 `json:"marketCap"` // j3: market cap real time.
	// MarketCapX float64 `json:"marketCap"`                   // j1: market cap (fallback when real time is N/A).
	// Currency   string  `json:"currency"`                    // String code for currency of stock.
	PreOpen    float64 `json:"preMarketChangePercent,omitempty"`
	AfterHours float64 `json:"postMarketChangePercent,omitempty"`
}

func FetchMarket() (*[]Quote, error) {
	return FetchQuotes(marketTickers)
}

// retrieve quotes for all tickers
func FetchQuotes(tickers []string) (*[]Quote, error) {
	result := []Quote{}
	if len(tickers) == 0 {
		return &result, nil
	}

	// fmt.Printf("Fetching quotes %v", tickers)
	url := fmt.Sprintf(apiURLv7, strings.Join(tickers, `,`))

	response, err := http.Get(url + apiURLv7ExtraParams)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result, err = unmarshalQuotes(body)
	if err != nil {
		return nil, err
	}

	return &result, nil
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
		fmt.Printf("body: %s", string(body))
		return nil, err
	}

	results := q["quoteResponse"]["result"]

	return results, nil
}

func float2Str(v float64, precision int) string {
	unit := ""
	switch {
	case v > 1.0e12:
		v = v / 1.0e12
		unit = "T"
	case v > 1.0e9:
		v = v / 1.0e9
		unit = "B"
	case v > 1.0e6:
		v = v / 1.0e6
		unit = "M"
	case v > 1.0e5:
		v = v / 1.0e3
		unit = "K"
	default:
		unit = ""
	}
	// parse
	return fmt.Sprintf("%0.*f%s", precision, v, unit)
}
