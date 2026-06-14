package stocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/localitas/localitas-go"
)

type handler struct {
	app *App
}

func (h *handler) handleListPortfolios(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	portfolios, err := h.app.Store.ListPortfolios(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, portfolios)
}

func (h *handler) handleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	userID := client.UserIDFromRequest(r)
	p, err := h.app.Store.CreatePortfolio(r.Context(), userID, req.Name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *handler) handleUpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	h.app.Store.UpdatePortfolio(r.Context(), id, req.Name)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleDeletePortfolio(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.app.Store.DeletePortfolio(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleListHoldings(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		writeErr(w, http.StatusBadRequest, "portfolio_id is required")
		return
	}
	holdings, err := h.app.Store.ListHoldings(r.Context(), portfolioID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
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
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) handleAddHolding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string `json:"portfolio_id"`
		Symbol      string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.PortfolioID == "" || req.Symbol == "" {
		writeErr(w, http.StatusBadRequest, "portfolio_id and symbol are required")
		return
	}
	req.Symbol = strings.ToUpper(req.Symbol)
	hold, err := h.app.Store.AddHolding(r.Context(), req.PortfolioID, req.Symbol)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusCreated, hold)
}

func (h *handler) handleUpdateHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		AllocationPct float64 `json:"allocation_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	h.app.Store.UpdateHolding(r.Context(), id, req.AllocationPct)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleReorderHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Position int64 `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	h.app.Store.UpdateSortPosition(r.Context(), id, req.Position)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleDeleteHolding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.app.Store.DeleteHolding(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleQuote(w http.ResponseWriter, r *http.Request) {
	symbols := r.URL.Query().Get("symbols")
	if symbols == "" {
		writeErr(w, http.StatusBadRequest, "symbols is required")
		return
	}
	quotes, err := FetchQuotes(strings.Split(symbols, ","))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, quotes)
}

func (h *handler) handleChart(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	rangeStr := r.URL.Query().Get("range")
	if symbol == "" {
		writeErr(w, http.StatusBadRequest, "symbol is required")
		return
	}
	if rangeStr == "" {
		rangeStr = "1d"
	}

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
			writeErr(w, http.StatusInternalServerError, "%v", err)
			return
		}
		writeJSON(w, http.StatusOK, points)
		return
	}

	cached, _ := h.app.Store.GetCachedChart(r.Context(), symbol, cacheInterval)

	if len(cached) == 0 {
		points, err := FetchChart(symbol, rangeStr, r.URL.Query().Get("interval"))
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "%v", err)
			return
		}
		h.app.Store.SaveChartPoints(r.Context(), symbol, cacheInterval, points)
		writeJSON(w, http.StatusOK, points)
		return
	}

	lastTS := h.app.Store.GetLastCachedTimestamp(r.Context(), symbol, cacheInterval)
	fresh, _ := FetchChart(symbol, "5d", cacheInterval)
	var newPoints []ChartPoint
	for _, p := range fresh {
		if p.Timestamp > lastTS {
			newPoints = append(newPoints, p)
		}
	}
	if len(newPoints) > 0 {
		h.app.Store.SaveChartPoints(r.Context(), symbol, cacheInterval, newPoints)
		cached = append(cached, newPoints...)
	}

	writeJSON(w, http.StatusOK, cached)
}

func (h *handler) handleETFHoldings(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedETFHoldings(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if len(cached) > 0 && updatedAt > thirtyDaysAgo {
		writeJSON(w, http.StatusOK, cached)
		return
	}

	holdings, err := FetchETFHoldings(symbol)
	if err != nil {
		if len(cached) > 0 {
			writeJSON(w, http.StatusOK, cached)
			return
		}
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveETFHoldings(r.Context(), symbol, holdings)
	writeJSON(w, http.StatusOK, holdings)
}

func (h *handler) handleEarnings(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, http.StatusBadRequest, "symbol is required")
		return
	}
	events, err := FetchEarnings([]string{symbol})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *handler) handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string             `json:"portfolio_id"`
		Amount      float64            `json:"amount"`
		Range       string             `json:"range"`
		Allocations map[string]float64 `json:"allocations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Amount <= 0 || len(req.Allocations) == 0 {
		writeErr(w, http.StatusBadRequest, "amount and allocations are required")
		return
	}
	if req.Range == "" {
		req.Range = "1y"
	}

	result := SimulationResult{TotalInvested: req.Amount}

	for symbol, pct := range req.Allocations {
		invested := req.Amount * (pct / 100)
		points, err := FetchChart(symbol, req.Range, "")
		if err != nil || len(points) == 0 {
			continue
		}
		startPrice := points[0].Close
		endPrice := points[len(points)-1].Close
		if startPrice <= 0 {
			continue
		}
		shares := invested / startPrice
		value := shares * endPrice
		gain := value - invested
		gainPct := 0.0
		if invested > 0 {
			gainPct = (gain / invested) * 100
		}
		sh := SimulationHolding{
			Symbol:        symbol,
			AllocationPct: pct,
			Invested:      invested,
			StartPrice:    startPrice,
			EndPrice:      endPrice,
			Shares:        shares,
			Value:         value,
			Gain:          gain,
			GainPct:       gainPct,
		}
		result.Holdings = append(result.Holdings, sh)
		result.TotalValue += value
	}

	result.TotalGain = result.TotalValue - result.TotalInvested
	if result.TotalInvested > 0 {
		result.TotalGainPct = (result.TotalGain / result.TotalInvested) * 100
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *handler) handleFinancials(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedFinancials(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if len(cached) > 0 && updatedAt > thirtyDaysAgo {
		writeJSON(w, http.StatusOK, cached)
		return
	}

	stmts, err := FetchFinancials(symbol)
	if err != nil {
		if len(cached) > 0 {
			writeJSON(w, http.StatusOK, cached)
			return
		}
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveFinancials(r.Context(), stmts)
	writeJSON(w, http.StatusOK, stmts)
}

func (h *handler) handleAnalystTargets(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeErr(w, http.StatusBadRequest, "symbol is required")
		return
	}

	cached, updatedAt, _ := h.app.Store.GetCachedAnalyst(r.Context(), symbol)
	thirtyDaysAgo := time.Now().UTC().Unix() - 30*24*3600
	if cached != nil && updatedAt > thirtyDaysAgo {
		writeJSON(w, http.StatusOK, cached)
		return
	}

	targets, err := FetchAnalystTargets(symbol)
	if err != nil {
		if cached != nil {
			writeJSON(w, http.StatusOK, cached)
			return
		}
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	h.app.Store.SaveAnalyst(r.Context(), symbol, targets)
	writeJSON(w, http.StatusOK, targets)
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, http.StatusBadRequest, "q is required")
		return
	}
	results, err := SearchSymbol(q)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf(format, args...)})
}
