package stocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"
)

type chartWidgetData struct {
	UID           string
	Title         string
	Range         string
	MultiTicker   bool
	SymbolsJSON   template.JS
	ChartDataJSON template.JS
	TableRows     []chartTableRow
}

type chartTableRow struct {
	Date   string
	Symbol string
	Open   string
	High   string
	Low    string
	Close  string
	Volume string
}

var chartWidgetTmpl = template.Must(
	template.New("chart_widget").Parse(mustReadChartTemplate()),
)

func mustReadChartTemplate() string {
	data, err := TemplatesFS.ReadFile("templates/partials/_chart_widget.html")
	if err != nil {
		panic("failed to read chart widget template: " + err.Error())
	}
	s := string(data)
	s = strings.TrimPrefix(s, `{{define "chart_widget"}}`)
	s = strings.TrimSuffix(strings.TrimSpace(s), `{{end}}`)
	return s
}

func renderChartWidget(symbols []string, allData map[string][]ChartPoint, rangeStr string) string {
	if len(allData) == 0 {
		return `<p style="color:var(--color-text-secondary);font-size:0.875rem;">No chart data available.</p>`
	}

	uid := fmt.Sprintf("stk-%d", time.Now().UnixNano()%100000)
	title := strings.Join(symbols, ", ")
	multiTicker := len(symbols) > 1

	symbolsJSON, _ := json.Marshal(symbols)
	chartDataJSON, _ := json.Marshal(allData)

	// Build table rows grouped by date ascending
	type dateRow struct {
		ts     int64
		symbol string
		point  ChartPoint
	}
	var rows []dateRow
	for sym, points := range allData {
		for _, p := range points {
			rows = append(rows, dateRow{ts: p.Timestamp, symbol: sym, point: p})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ts == rows[j].ts {
			return rows[i].symbol < rows[j].symbol
		}
		return rows[i].ts < rows[j].ts
	})

	var tableRows []chartTableRow
	for _, r := range rows {
		tableRows = append(tableRows, chartTableRow{
			Date:   time.Unix(r.ts, 0).Format("2006-01-02"),
			Symbol: r.symbol,
			Open:   fmt.Sprintf("%.2f", r.point.Open),
			High:   fmt.Sprintf("%.2f", r.point.High),
			Low:    fmt.Sprintf("%.2f", r.point.Low),
			Close:  fmt.Sprintf("%.2f", r.point.Close),
			Volume: formatVolume(r.point.Volume),
		})
	}

	data := chartWidgetData{
		UID:           uid,
		Title:         title,
		Range:         rangeStr,
		MultiTicker:   multiTicker,
		SymbolsJSON:   template.JS(symbolsJSON),
		ChartDataJSON: template.JS(chartDataJSON),
		TableRows:     tableRows,
	}

	var buf bytes.Buffer
	if err := chartWidgetTmpl.Execute(&buf, data); err != nil {
		return `<p style="color:var(--color-error);">Failed to render chart.</p>`
	}
	return buf.String()
}

func formatVolume(v int64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(v)/1_000_000_000)
	}
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(v)/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(v)/1_000)
	}
	return fmt.Sprintf("%d", v)
}
