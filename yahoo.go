package stocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const yahooChartURL = "https://query1.finance.yahoo.com/v8/finance/chart"

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Jar:     nil,
}

var yahooCrumb string
var yahooCookies []*http.Cookie

func ensureYahooCrumb() error {
	if yahooCrumb != "" {
		return nil
	}
	req, _ := http.NewRequest("GET", "https://fc.yahoo.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	yahooCookies = resp.Cookies()
	resp.Body.Close()

	req2, _ := http.NewRequest("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	for _, c := range yahooCookies {
		req2.AddCookie(c)
	}
	resp2, err := httpClient.Do(req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	body, _ := io.ReadAll(resp2.Body)
	yahooCrumb = string(body)
	return nil
}

func yahooAuthGet(u string) ([]byte, error) {
	if err := ensureYahooCrumb(); err != nil {
		return nil, err
	}
	sep := "?"
	if len(u) > 0 && (u[len(u)-1] != '?' && !contains(u, "?")) {
		sep = "?"
	} else {
		sep = "&"
	}
	fullURL := u + sep + "crumb=" + url.QueryEscape(yahooCrumb)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	for _, c := range yahooCookies {
		req.AddCookie(c)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 {
		yahooCrumb = ""
		return yahooAuthGet(u)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	return body, nil
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func yahooGet(u string) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	return body, nil
}

type chartMeta struct {
	Symbol               string  `json:"symbol"`
	ShortName            string  `json:"shortName"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	PreviousClose        float64 `json:"previousClose"`
	ChartPreviousClose   float64 `json:"chartPreviousClose"`
	ExchangeName         string  `json:"exchangeName"`
	FullExchangeName     string  `json:"fullExchangeName"`
	Currency             string  `json:"currency"`
	InstrumentType       string  `json:"instrumentType"`
	FiftyTwoWeekHigh     float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      float64 `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
	RegularMarketVolume  int64   `json:"regularMarketVolume"`
}

type chartResponse struct {
	Chart struct {
		Result []struct {
			Meta       chartMeta `json:"meta"`
			Timestamp  []int64   `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []interface{} `json:"open"`
					High   []interface{} `json:"high"`
					Low    []interface{} `json:"low"`
					Close  []interface{} `json:"close"`
					Volume []interface{} `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

func FetchQuotes(symbols []string) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}
	quotes := make([]Quote, 0, len(symbols))
	for _, sym := range symbols {
		q, err := fetchQuoteFromChart(sym)
		if err != nil {
			continue
		}
		quotes = append(quotes, *q)
	}
	return quotes, nil
}

func fetchQuoteFromChart(symbol string) (*Quote, error) {
	body, err := yahooGet(fmt.Sprintf("%s/%s?range=1d&interval=5m&includePrePost=true", yahooChartURL, url.PathEscape(symbol)))
	if err != nil {
		return nil, err
	}

	var cr chartResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, err
	}
	if len(cr.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	m := cr.Chart.Result[0].Meta
	prevClose := m.ChartPreviousClose
	if prevClose == 0 {
		prevClose = m.PreviousClose
	}
	change := m.RegularMarketPrice - prevClose
	changePct := 0.0
	if prevClose > 0 {
		changePct = (change / prevClose) * 100
	}

	quoteType := "EQUITY"
	if m.InstrumentType == "ETF" {
		quoteType = "ETF"
	}

	return &Quote{
		Symbol:        m.Symbol,
		Name:          m.ShortName,
		Price:         m.RegularMarketPrice,
		Change:        change,
		ChangePercent: changePct,
		PreviousClose: prevClose,
		DayHigh:       m.RegularMarketDayHigh,
		DayLow:        m.RegularMarketDayLow,
		Volume:        m.RegularMarketVolume,
		Week52High:    m.FiftyTwoWeekHigh,
		Week52Low:     m.FiftyTwoWeekLow,
		Exchange:      m.FullExchangeName,
		QuoteType:     quoteType,
		Currency:      m.Currency,
	}, nil
}

func FetchChart(symbol, rangeStr, interval string) ([]ChartPoint, error) {
	if interval == "" {
		switch rangeStr {
		case "1d":
			interval = "5m"
		case "5d":
			interval = "15m"
		case "1mo":
			interval = "1h"
		case "ytd":
			interval = "1d"
		case "max":
			interval = "1wk"
		default:
			interval = "1d"
		}
	}

	body, err := yahooGet(fmt.Sprintf("%s/%s?range=%s&interval=%s&includePrePost=true", yahooChartURL, url.PathEscape(symbol), rangeStr, interval))
	if err != nil {
		return nil, err
	}

	var cr chartResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, err
	}
	if len(cr.Chart.Result) == 0 || len(cr.Chart.Result[0].Indicators.Quote) == 0 {
		return nil, nil
	}

	r := cr.Chart.Result[0]
	q := r.Indicators.Quote[0]
	points := make([]ChartPoint, 0, len(r.Timestamp))
	for i, ts := range r.Timestamp {
		cp := ChartPoint{Timestamp: ts}
		if i < len(q.Close) && q.Close[i] != nil {
			if v, ok := q.Close[i].(float64); ok {
				cp.Close = v
			}
		}
		if i < len(q.Open) && q.Open[i] != nil {
			if v, ok := q.Open[i].(float64); ok {
				cp.Open = v
			}
		}
		if i < len(q.High) && q.High[i] != nil {
			if v, ok := q.High[i].(float64); ok {
				cp.High = v
			}
		}
		if i < len(q.Low) && q.Low[i] != nil {
			if v, ok := q.Low[i].(float64); ok {
				cp.Low = v
			}
		}
		if i < len(q.Volume) && q.Volume[i] != nil {
			if v, ok := q.Volume[i].(float64); ok {
				cp.Volume = int64(v)
			}
		}
		if cp.Close > 0 {
			points = append(points, cp)
		}
	}
	return points, nil
}

func FetchETFHoldings(symbol string) ([]ETFHolding, error) {
	body, err := yahooAuthGet(fmt.Sprintf("https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=topHoldings", url.PathEscape(symbol)))
	if err != nil {
		return nil, err
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				TopHoldings struct {
					Holdings []struct {
						Symbol         string `json:"symbol"`
						HoldingName    string `json:"holdingName"`
						HoldingPercent struct {
							Raw float64 `json:"raw"`
						} `json:"holdingPercent"`
					} `json:"holdings"`
				} `json:"topHoldings"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.QuoteSummary.Result) == 0 {
		return nil, nil
	}

	holdings := make([]ETFHolding, 0)
	for _, h := range result.QuoteSummary.Result[0].TopHoldings.Holdings {
		holdings = append(holdings, ETFHolding{
			Symbol: h.Symbol,
			Name:   h.HoldingName,
			Weight: h.HoldingPercent.Raw * 100,
		})
	}
	return holdings, nil
}

func FetchEarnings(symbols []string) ([]EarningsEvent, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	body, err := yahooAuthGet(fmt.Sprintf("https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=calendarEvents", url.PathEscape(symbols[0])))
	if err != nil {
		return nil, nil
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				CalendarEvents struct {
					Earnings struct {
						EarningsDate []struct {
							Fmt string `json:"fmt"`
						} `json:"earningsDate"`
						EarningsAverage struct {
							Fmt string `json:"fmt"`
						} `json:"earningsAverage"`
					} `json:"earnings"`
				} `json:"calendarEvents"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil
	}

	events := make([]EarningsEvent, 0)
	if len(result.QuoteSummary.Result) > 0 {
		ce := result.QuoteSummary.Result[0].CalendarEvents
		for _, ed := range ce.Earnings.EarningsDate {
			events = append(events, EarningsEvent{
				Symbol:      symbols[0],
				Date:        ed.Fmt,
				EpsEstimate: ce.Earnings.EarningsAverage.Fmt,
			})
		}
	}
	return events, nil
}

func FetchAnalystTargets(symbol string) (*AnalystTargets, error) {
	body, err := yahooAuthGet(fmt.Sprintf("https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=financialData", url.PathEscape(symbol)))
	if err != nil {
		return nil, err
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				FinancialData struct {
					TargetLowPrice          struct{ Raw float64 } `json:"targetLowPrice"`
					TargetMeanPrice         struct{ Raw float64 } `json:"targetMeanPrice"`
					TargetMedianPrice       struct{ Raw float64 } `json:"targetMedianPrice"`
					TargetHighPrice         struct{ Raw float64 } `json:"targetHighPrice"`
					NumberOfAnalystOpinions struct{ Raw int }     `json:"numberOfAnalystOpinions"`
					RecommendationKey       string                `json:"recommendationKey"`
				} `json:"financialData"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.QuoteSummary.Result) == 0 {
		return nil, nil
	}

	fd := result.QuoteSummary.Result[0].FinancialData
	return &AnalystTargets{
		Low:              fd.TargetLowPrice.Raw,
		Mean:             fd.TargetMeanPrice.Raw,
		Median:           fd.TargetMedianPrice.Raw,
		High:             fd.TargetHighPrice.Raw,
		NumberOfAnalysts: fd.NumberOfAnalystOpinions.Raw,
		Recommendation:   fd.RecommendationKey,
	}, nil
}

func FetchFinancials(symbol string) ([]FinancialStatement, error) {
	body, err := yahooAuthGet(fmt.Sprintf("https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=incomeStatementHistory,incomeStatementHistoryQuarterly,earningsHistory", url.PathEscape(symbol)))
	if err != nil {
		return nil, err
	}

	type rawVal struct {
		Raw int64  `json:"raw"`
		Fmt string `json:"fmt"`
	}
	type rawFloat struct {
		Raw float64 `json:"raw"`
		Fmt string  `json:"fmt"`
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				IncomeStatementHistory struct {
					IncomeStatementHistory []struct {
						EndDate         rawVal `json:"endDate"`
						TotalRevenue    rawVal `json:"totalRevenue"`
						NetIncome       rawVal `json:"netIncome"`
						GrossProfit     rawVal `json:"grossProfit"`
						OperatingIncome rawVal `json:"operatingIncome"`
					} `json:"incomeStatementHistory"`
				} `json:"incomeStatementHistory"`
				IncomeStatementHistoryQuarterly struct {
					IncomeStatementHistory []struct {
						EndDate         rawVal `json:"endDate"`
						TotalRevenue    rawVal `json:"totalRevenue"`
						NetIncome       rawVal `json:"netIncome"`
						GrossProfit     rawVal `json:"grossProfit"`
						OperatingIncome rawVal `json:"operatingIncome"`
					} `json:"incomeStatementHistory"`
				} `json:"incomeStatementHistoryQuarterly"`
				EarningsHistory struct {
					History []struct {
						Quarter         rawVal   `json:"quarter"`
						EpsEstimate     rawFloat `json:"epsEstimate"`
						EpsActual       rawFloat `json:"epsActual"`
						SurprisePercent rawFloat `json:"surprisePercent"`
					} `json:"history"`
				} `json:"earningsHistory"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.QuoteSummary.Result) == 0 {
		return nil, nil
	}

	r := result.QuoteSummary.Result[0]
	var stmts []FinancialStatement

	for _, s := range r.IncomeStatementHistory.IncomeStatementHistory {
		stmts = append(stmts, FinancialStatement{
			Symbol:          symbol,
			PeriodType:      "annual",
			EndDate:         s.EndDate.Fmt,
			TotalRevenue:    s.TotalRevenue.Raw,
			NetIncome:       s.NetIncome.Raw,
			GrossProfit:     s.GrossProfit.Raw,
			OperatingIncome: s.OperatingIncome.Raw,
		})
	}

	for _, s := range r.IncomeStatementHistoryQuarterly.IncomeStatementHistory {
		stmts = append(stmts, FinancialStatement{
			Symbol:          symbol,
			PeriodType:      "quarterly",
			EndDate:         s.EndDate.Fmt,
			TotalRevenue:    s.TotalRevenue.Raw,
			NetIncome:       s.NetIncome.Raw,
			GrossProfit:     s.GrossProfit.Raw,
			OperatingIncome: s.OperatingIncome.Raw,
		})
	}

	for _, e := range r.EarningsHistory.History {
		stmts = append(stmts, FinancialStatement{
			Symbol:         symbol,
			PeriodType:     "earnings",
			EndDate:        e.Quarter.Fmt,
			EpsEstimate:    e.EpsEstimate.Raw,
			EpsActual:      e.EpsActual.Raw,
			EpsSurprisePct: e.SurprisePercent.Raw,
		})
	}

	for i := 1; i < len(stmts); i++ {
		if stmts[i].PeriodType == stmts[i-1].PeriodType && stmts[i-1].TotalRevenue > 0 && stmts[i].TotalRevenue > 0 {
			stmts[i].RevenueGrowth = float64(stmts[i-1].TotalRevenue-stmts[i].TotalRevenue) / float64(stmts[i].TotalRevenue) * 100
		}
	}

	return stmts, nil
}

func SearchSymbol(query string) ([]Quote, error) {
	body, err := yahooGet(fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=8&newsCount=0", url.QueryEscape(query)))
	if err != nil {
		return nil, nil
	}

	var result struct {
		Quotes []struct {
			Symbol    string `json:"symbol"`
			ShortName string `json:"shortname"`
			QuoteType string `json:"quoteType"`
			Exchange  string `json:"exchange"`
		} `json:"quotes"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil
	}

	quotes := make([]Quote, 0, len(result.Quotes))
	for _, q := range result.Quotes {
		quotes = append(quotes, Quote{
			Symbol:    q.Symbol,
			Name:      q.ShortName,
			QuoteType: q.QuoteType,
			Exchange:  q.Exchange,
		})
	}
	return quotes, nil
}
