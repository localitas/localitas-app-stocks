package stocks

import (
	"net/http"

	client "github.com/localitas/localitas-go"
)

var StocksAPIDoc = client.APISpec{
	AppName:     "Stocks",
	Version:     "1.0.0",
	Description: "Stock portfolio tracker with Yahoo Finance data. Supports real-time quotes, historical charts, ETF holdings, analyst targets, financial statements, and portfolio simulation.",
	Keywords:    []string{"stocks", "shares", "portfolio", "market", "trading", "equity", "ticker", "quote", "price", "dividend", "ETF", "investment", "finance", "S&P", "NASDAQ", "NYSE"},
	Endpoints: []client.APIEndpoint{
		{Method: "GET", Path: "/api/quote", Summary: "Get real-time quotes", QueryParams: []client.APIParam{{Name: "symbols", Type: "string", Required: true, Description: "Comma-separated symbols (AAPL,MSFT)"}}, Response: &client.APIBody{ContentType: "application/json", Example: `[{"symbol":"AAPL","name":"Apple Inc","price":210.50,"change":2.30,"change_percent":1.10}]`}},
		{Method: "GET", Path: "/api/chart", Summary: "Get price chart data", QueryParams: []client.APIParam{{Name: "symbol", Type: "string", Required: true, Description: "Stock symbol"}, {Name: "range", Type: "string", Required: false, Description: "1d, 5d, 1mo, 3mo, ytd, 1y, 5y, max"}}, Response: &client.APIBody{ContentType: "application/json", Example: `[{"timestamp":1714400000,"close":210.50,"volume":50000000}]`}},
		{Method: "GET", Path: "/api/search", Summary: "Search for symbols", QueryParams: []client.APIParam{{Name: "q", Type: "string", Required: true, Description: "Search query"}}},
		{Method: "GET", Path: "/api/etf-holdings", Summary: "Get ETF top holdings", QueryParams: []client.APIParam{{Name: "symbol", Type: "string", Required: true, Description: "ETF symbol (QQQ, SPY)"}}},
		{Method: "GET", Path: "/api/analyst-targets", Summary: "Get analyst price targets", QueryParams: []client.APIParam{{Name: "symbol", Type: "string", Required: true, Description: "Stock symbol"}}},
		{Method: "GET", Path: "/api/financials", Summary: "Get income statements and earnings", QueryParams: []client.APIParam{{Name: "symbol", Type: "string", Required: true, Description: "Stock symbol"}}},
		{Method: "GET", Path: "/api/earnings", Summary: "Get upcoming earnings dates", QueryParams: []client.APIParam{{Name: "symbol", Type: "string", Required: true, Description: "Stock symbol"}}},
		{Method: "GET", Path: "/api/portfolios", Summary: "List all portfolios"},
		{Method: "POST", Path: "/api/portfolios", Summary: "Create portfolio", RequestBody: &client.APIBody{ContentType: "application/json", Example: `{"name":"Tech Stocks"}`}},
		{Method: "PUT", Path: "/api/portfolios/{id}", Summary: "Update portfolio name"},
		{Method: "DELETE", Path: "/api/portfolios/{id}", Summary: "Delete portfolio"},
		{Method: "GET", Path: "/api/holdings", Summary: "List holdings with quotes", QueryParams: []client.APIParam{{Name: "portfolio_id", Type: "string", Required: true, Description: "Portfolio ID"}}},
		{Method: "POST", Path: "/api/holdings", Summary: "Add symbol to portfolio", RequestBody: &client.APIBody{ContentType: "application/json", Example: `{"portfolio_id":"abc","symbol":"AAPL"}`}},
		{Method: "PUT", Path: "/api/holdings/{id}", Summary: "Update allocation %", RequestBody: &client.APIBody{ContentType: "application/json", Example: `{"allocation_pct":25.0}`}},
		{Method: "DELETE", Path: "/api/holdings/{id}", Summary: "Remove holding"},
		{Method: "POST", Path: "/api/simulate", Summary: "Run portfolio simulation", RequestBody: &client.APIBody{ContentType: "application/json", Example: `{"portfolio_id":"abc","amount":10000,"range":"1y","allocations":{"AAPL":50,"MSFT":50}}`}},
		{Method: "POST", Path: "/api/refresh", Summary: "Refresh stock quotes", RequestBody: &client.APIBody{ContentType: "application/json", Example: `{"portfolio_id":"abc"}`}, Response: &client.APIBody{ContentType: "application/json", Example: `{"refreshed":1,"symbols":["AAPL"],"quotes":[{"symbol":"AAPL","name":"Apple Inc","price":210.50,"change":2.30,"change_percent":1.10}]}`}},
	},
}

func HandleSwagger(w http.ResponseWriter, r *http.Request) {
	StocksAPIDoc.ServeHTTP(w, r)
}
