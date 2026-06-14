package stocks

import "time"

type Portfolio struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Holding struct {
	ID            string    `json:"id"`
	PortfolioID   string    `json:"portfolio_id"`
	Symbol        string    `json:"symbol"`
	AllocationPct float64   `json:"allocation_pct"`
	SortPosition  int64     `json:"sort_position"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Quote struct {
	Symbol           string  `json:"symbol"`
	Name             string  `json:"name"`
	Price            float64 `json:"price"`
	Change           float64 `json:"change"`
	ChangePercent    float64 `json:"change_percent"`
	PreviousClose    float64 `json:"previous_close"`
	Open             float64 `json:"open"`
	DayHigh          float64 `json:"day_high"`
	DayLow           float64 `json:"day_low"`
	Volume           int64   `json:"volume"`
	AvgVolume        int64   `json:"avg_volume"`
	MarketCap        int64   `json:"market_cap"`
	PERatio          float64 `json:"pe_ratio"`
	EPS              float64 `json:"eps"`
	Week52High       float64 `json:"week_52_high"`
	Week52Low        float64 `json:"week_52_low"`
	DividendYield    float64 `json:"dividend_yield"`
	PreMarketPrice   float64 `json:"pre_market_price"`
	PreMarketChange  float64 `json:"pre_market_change"`
	PostMarketPrice  float64 `json:"post_market_price"`
	PostMarketChange float64 `json:"post_market_change"`
	MarketState      string  `json:"market_state"`
	Exchange         string  `json:"exchange"`
	QuoteType        string  `json:"quote_type"`
	Currency         string  `json:"currency"`
}

type ChartPoint struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

type ETFHolding struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type EarningsEvent struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Date        string `json:"date"`
	EpsEstimate string `json:"eps_estimate"`
	EpsActual   string `json:"eps_actual"`
	Surprise    string `json:"surprise"`
}

type HoldingWithQuote struct {
	Holding
	Quote *Quote `json:"quote,omitempty"`
}

type AnalystTargets struct {
	Low              float64 `json:"low"`
	Mean             float64 `json:"mean"`
	Median           float64 `json:"median"`
	High             float64 `json:"high"`
	NumberOfAnalysts int     `json:"number_of_analysts"`
	Recommendation   string  `json:"recommendation"`
}

type FinancialStatement struct {
	Symbol          string  `json:"symbol"`
	PeriodType      string  `json:"period_type"`
	EndDate         string  `json:"end_date"`
	TotalRevenue    int64   `json:"total_revenue"`
	NetIncome       int64   `json:"net_income"`
	GrossProfit     int64   `json:"gross_profit"`
	OperatingIncome int64   `json:"operating_income"`
	RevenueGrowth   float64 `json:"revenue_growth,omitempty"`
	EpsEstimate     float64 `json:"eps_estimate,omitempty"`
	EpsActual       float64 `json:"eps_actual,omitempty"`
	EpsSurprisePct  float64 `json:"eps_surprise_pct,omitempty"`
}

type SimulationResult struct {
	TotalInvested float64             `json:"total_invested"`
	TotalValue    float64             `json:"total_value"`
	TotalGain     float64             `json:"total_gain"`
	TotalGainPct  float64             `json:"total_gain_pct"`
	Holdings      []SimulationHolding `json:"holdings"`
}

type SimulationHolding struct {
	Symbol        string  `json:"symbol"`
	AllocationPct float64 `json:"allocation_pct"`
	Invested      float64 `json:"invested"`
	StartPrice    float64 `json:"start_price"`
	EndPrice      float64 `json:"end_price"`
	Shares        float64 `json:"shares"`
	Value         float64 `json:"value"`
	Gain          float64 `json:"gain"`
	GainPct       float64 `json:"gain_pct"`
}
