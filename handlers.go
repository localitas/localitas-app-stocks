package stocks

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/localitas/localitas-go"
	"github.com/localitas/localitas-go/httputil"
)

type handler struct {
	app *App
}

func (h *handler) handleListPortfolios(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	portfolios, err := h.app.Store.ListPortfolios(r.Context(), userID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusOK, portfolios)
}

func (h *handler) handleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		writeErr(w, r, http.StatusBadRequest, "name is required")
		return
	}
	userID := client.UserIDFromRequest(r)
	p, err := h.app.Store.CreatePortfolio(r.Context(), userID, req.Name)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusCreated, p)
}

func (h *handler) handleUpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	h.app.Store.UpdatePortfolio(r.Context(), id, req.Name)
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleDeletePortfolio(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.app.Store.DeletePortfolio(r.Context(), id)
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleListHoldings(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		writeErr(w, r, http.StatusBadRequest, "portfolio_id is required")
		return
	}
	holdings, err := h.app.Store.ListHoldings(r.Context(), portfolioID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}

	symbols := make([]string, 0, len(holdings))
	for _, hold := range holdings {
		symbols = append(symbols, hold.Symbol)
	}

	quotes, _ := FetchQuotes(symbols)
	quoteMap := make(map[string]*Quote)
	for i := range quotes {
		quoteMap[quotes[i].Symbol] = &quotes[i]
	}

	result := make([]HoldingWithQuote, 0, len(holdings))
	for _, hold := range holdings {
		hwq := HoldingWithQuote{Holding: *hold}
		if q, ok := quoteMap[hold.Symbol]; ok {
			hwq.Quote = q
		}
		result = append(result, hwq)
	}
	writeJSON(w, r, http.StatusOK, result)
}

func (h *handler) handleAddHolding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string `json:"portfolio_id"`
		Symbol      string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if req.PortfolioID == "" || req.Symbol == "" {
		writeErr(w, r, http.StatusBadRequest, "portfolio_id and symbol are required")
		return
	}
	req.Symbol = strings.ToUpper(req.Symbol)
	hold, err := h.app.Store.AddHolding(r.Context(), req.PortfolioID, req.Symbol)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusCreated, hold)
}

func (h *handler) handleUpdateHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		AllocationPct *float64 `json:"allocation_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if req.AllocationPct == nil {
		writeErr(w, r, http.StatusBadRequest, "allocation_pct is required")
		return
	}
	h.app.Store.UpdateHolding(r.Context(), id, *req.AllocationPct)
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleReorderHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Position int64 `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	h.app.Store.UpdateSortPosition(r.Context(), id, req.Position)
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleDeleteHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.app.Store.DeleteHolding(r.Context(), id)
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleQuote(w http.ResponseWriter, r *http.Request) {
	symbols := r.URL.Query().Get("symbols")
	if symbols == "" {
		writeErr(w, r, http.StatusBadRequest, "symbols is required")
		return
	}
	quotes, err := FetchQuotes(strings.Split(symbols, ","))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusOK, quotes)
}

func (h *handler) handleChart(w http.ResponseWriter, r *http.Request) {
	symbolStr := r.URL.Query().Get("symbol")
	rangeStr := r.URL.Query().Get("range")
	if symbolStr == "" {
		writeErr(w, r, http.StatusBadRequest, "symbol is required")
		return
	}
	if rangeStr == "" {
		rangeStr = "1mo"
	}

	symbols := strings.Split(symbolStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(strings.ToUpper(symbols[i]))
	}

	allData := make(map[string][]ChartPoint)
	for _, symbol := range symbols {
		points := h.fetchChartData(r, symbol, rangeStr)
		if len(points) > 0 {
			allData[symbol] = points
		}
	}

	if strings.EqualFold(r.Header.Get("Caller"), "llm") {
		w.Header().Set("Content-Type", "text/markdown")
		w.Write([]byte(renderChartWidget(symbols, allData, rangeStr)))
		return
	}

	if len(symbols) == 1 {
		writeJSON(w, r, http.StatusOK, allData[symbols[0]])
	} else {
		writeJSON(w, r, http.StatusOK, allData)
	}
}

func (h *handler) fetchChartData(r *http.Request, symbol, rangeStr string) []ChartPoint {
	cacheInterval := "1d"
	switch rangeStr {
	case "1d":
		cacheInterval = "5m"
	case "5d":
		cacheInterval = "15m"
	case "1mo":
		cacheInterval = "1h"
	case "max":
		cacheInterval = "1wk"
	}

	if rangeStr == "1d" || rangeStr == "5d" || rangeStr == "ytd" {
		points, err := FetchChart(symbol, rangeStr, r.URL.Query().Get("interval"))
		if err != nil {
			return nil
		}
		return points
	}

	cached, _ := h.app.Store.GetCachedChart(r.Context(), symbol, cacheInterval)

	if len(cached) == 0 {
		points, err := FetchChart(symbol, rangeStr, r.URL.Query().Get("interval"))
		if err != nil {
			return nil
		}
		h.app.Store.SaveChartPoints(r.Context(), symbol, cacheInterval, points)
		return points
	}

	lastTS := h.app.Store.GetLastCachedTimestamp(r.Context(), symbol, cacheInterval)
	fresh, _ := FetchChart(symbol, "5d", cacheInterval)
	for _, p := range fresh {
		if p.Timestamp > lastTS {
			cached = append(cached, p)
			h.app.Store.SaveChartPoints(r.Context(), symbol, cacheInterval, []ChartPoint{p})
		}
	}

	return cached
}

func (h *handler) handleETFHoldings(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, r, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedETFHoldings(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if len(cached) > 0 && updatedAt > thirtyDaysAgo {
		writeJSON(w, r, http.StatusOK, cached)
		return
	}

	holdings, err := FetchETFHoldings(symbol)
	if err != nil {
		if len(cached) > 0 {
			writeJSON(w, r, http.StatusOK, cached)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveETFHoldings(r.Context(), symbol, holdings)
	writeJSON(w, r, http.StatusOK, holdings)
}

func (h *handler) handleEarnings(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, r, http.StatusBadRequest, "symbol is required")
		return
	}
	events, err := FetchEarnings([]string{symbol})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusOK, events)
}

func (h *handler) handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string             `json:"portfolio_id"`
		Amount      float64            `json:"amount"`
		Window      string             `json:"window"`
		Allocations map[string]float64 `json:"allocations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Amount <= 0 || len(req.Allocations) == 0 {
		writeErr(w, r, http.StatusBadRequest, "amount and allocations are required")
		return
	}
	if req.Window == "" {
		req.Window = "1y"
	}

	windowYears := windowToYears(req.Window)

	result := SimulationResult{
		TotalInvested: req.Amount,
		Window:        req.Window,
	}

	var weightedReturn float64
	var totalWeight float64

	yahooRange := windowToYahooRange(req.Window)
	windowDays := int(windowYears * 365)

	for symbol, pct := range req.Allocations {
		invested := req.Amount * (pct / 100)
		points, err := FetchChart(symbol, yahooRange, "")
		if err != nil || len(points) == 0 {
			continue
		}
		if len(points) > windowDays && windowDays > 0 {
			points = points[len(points)-windowDays:]
		}
		startPrice := points[0].Close
		endPrice := points[len(points)-1].Close
		if startPrice <= 0 {
			continue
		}

		totalReturn := (endPrice - startPrice) / startPrice
		annualized := annualizeReturn(totalReturn, windowYears)

		sh := SimulationHolding{
			Symbol:           symbol,
			AllocationPct:    pct,
			Invested:         invested,
			AnnualizedReturn: annualized * 100,
			CurrentPrice:     endPrice,
		}
		result.Holdings = append(result.Holdings, sh)

		weightedReturn += annualized * (pct / 100)
		totalWeight += pct / 100
	}

	if totalWeight > 0 {
		result.AnnualizedReturn = (weightedReturn / totalWeight) * 100
	}

	projectionLabels := []struct {
		Label string
		Years float64
	}{
		{"1 Year", 1},
		{"2 Years", 2},
		{"3 Years", 3},
		{"5 Years", 5},
		{"10 Years", 10},
		{"15 Years", 15},
		{"20 Years", 20},
		{"25 Years", 25},
		{"30 Years", 30},
		{"35 Years", 35},
		{"40 Years", 40},
		{"45 Years", 45},
		{"50 Years", 50},
	}

	portfolioAnnualized := weightedReturn / totalWeight
	for _, p := range projectionLabels {
		projected := req.Amount * math.Pow(1+portfolioAnnualized, p.Years)
		gain := projected - req.Amount
		gainPct := 0.0
		if req.Amount > 0 {
			gainPct = (gain / req.Amount) * 100
		}
		result.Projections = append(result.Projections, ProjectionSnapshot{
			Label:          p.Label,
			Years:          p.Years,
			ProjectedValue: projected,
			ProjectedGain:  gain,
			ProjectedPct:   gainPct,
		})
	}

	writeJSON(w, r, http.StatusOK, result)
}

func windowToYears(window string) float64 {
	switch window {
	case "200d":
		return 200.0 / 365.0
	case "3mo":
		return 0.25
	case "6mo":
		return 0.5
	case "1y":
		return 1.0
	case "2y":
		return 2.0
	case "3y":
		return 3.0
	case "5y":
		return 5.0
	default:
		return 1.0
	}
}

func windowToYahooRange(window string) string {
	switch window {
	case "200d":
		return "1y"
	case "3y":
		return "5y"
	default:
		return window
	}
}

func annualizeReturn(totalReturn, years float64) float64 {
	if years <= 0 || totalReturn <= -1 {
		return 0
	}
	return math.Pow(1+totalReturn, 1/years) - 1
}

func (h *handler) handleFinancials(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, r, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedFinancials(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if len(cached) > 0 && updatedAt > thirtyDaysAgo {
		writeJSON(w, r, http.StatusOK, cached)
		return
	}

	stmts, err := FetchFinancials(symbol)
	if err != nil {
		if len(cached) > 0 {
			writeJSON(w, r, http.StatusOK, cached)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveFinancials(r.Context(), stmts)
	writeJSON(w, r, http.StatusOK, stmts)
}

func (h *handler) handleAnalystTargets(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, r, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedAnalyst(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if cached != nil && updatedAt > thirtyDaysAgo {
		writeJSON(w, r, http.StatusOK, cached)
		return
	}

	targets, err := FetchAnalystTargets(symbol)
	if err != nil {
		if cached != nil {
			writeJSON(w, r, http.StatusOK, cached)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveAnalyst(r.Context(), symbol, targets)
	writeJSON(w, r, http.StatusOK, targets)
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, r, http.StatusBadRequest, "q is required")
		return
	}
	results, err := SearchSymbol(q)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, r, http.StatusOK, results)
}

func (h *handler) collectRefreshSymbols(ctx context.Context, portfolioID string, symbols []string) ([]string, error) {
	symbolSet := make(map[string]bool)

	if len(symbols) > 0 {
		for _, s := range symbols {
			symbolSet[strings.ToUpper(strings.TrimSpace(s))] = true
		}
	} else if portfolioID != "" {
		holdings, err := h.app.Store.ListHoldings(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		for _, hold := range holdings {
			symbolSet[hold.Symbol] = true
		}
	} else {
		portfolios, err := h.app.Store.ListAllPortfolios(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range portfolios {
			holdings, err := h.app.Store.ListHoldings(ctx, p.ID)
			if err != nil {
				continue
			}
			for _, hold := range holdings {
				symbolSet[hold.Symbol] = true
			}
		}
	}

	result := make([]string, 0, len(symbolSet))
	for s := range symbolSet {
		result = append(result, s)
	}
	return result, nil
}

func (h *handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string   `json:"portfolio_id"`
		Symbols     []string `json:"symbols"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	work := func(ctx context.Context) (map[string]interface{}, error) {
		symbols, err := h.collectRefreshSymbols(ctx, req.PortfolioID, req.Symbols)
		if err != nil {
			return nil, err
		}
		if len(symbols) == 0 {
			return map[string]interface{}{"refreshed": 0}, nil
		}
		quotes, err := FetchQuotes(symbols)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"refreshed": len(quotes),
			"symbols":   symbols,
		}, nil
	}

	if client.RunAsync(w, r, h.app.client, work) {
		return
	}

	symbols, err := h.collectRefreshSymbols(r.Context(), req.PortfolioID, req.Symbols)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "%v", err)
		return
	}

	if len(symbols) == 0 {
		writeJSON(w, r, http.StatusOK, RefreshResponse{Refreshed: 0})
		return
	}

	quotes, err := FetchQuotes(symbols)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "fetch quotes: %v", err)
		return
	}

	writeJSON(w, r, http.StatusOK, RefreshResponse{
		Refreshed: len(quotes),
		Symbols:   symbols,
		Quotes:    quotes,
	})
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, v interface{}) {
	httputil.WriteResponse(w, r, status, v)
}

func writeErr(w http.ResponseWriter, r *http.Request, status int, format string, args ...interface{}) {
	httputil.WriteError(w, r, status, format, args...)
}
