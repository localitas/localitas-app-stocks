package stocks

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

type App struct {
	Store    *Store
	BasePath string
	client   *client.Client
}

func New(c *client.Client, basePath string) *App {
	if basePath == "" {
		basePath = "/"
	}
	return &App{BasePath: basePath, client: c}
}

func (a *App) InitStore(coreURL, dbID, token string) error {
	store, err := OpenStore(coreURL, dbID, token)
	if err != nil {
		return err
	}
	a.Store = store
	return nil
}

func (a *App) Install(ctx context.Context) (string, error) {
	for attempt := 1; ; attempt++ {
		db, err := a.client.CreateSystemDatabase(ctx, DatabaseName)
		if err != nil {
			log.Printf("install: attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			log.Printf("install: migrations attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		log.Printf("stocks index template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	tmpl.ExecuteTemplate(w, "index.html", map[string]string{"BasePath": a.BasePath})
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	h := &handler{app: a}

	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /swagger.json", HandleSwagger)
	mux.HandleFunc("GET /help.md", handleHelpMarkdown)
	mux.HandleFunc("GET /api/portfolios", h.handleListPortfolios)
	mux.HandleFunc("POST /api/portfolios", h.handleCreatePortfolio)
	mux.HandleFunc("PUT /api/portfolios/{id}", h.handleUpdatePortfolio)
	mux.HandleFunc("DELETE /api/portfolios/{id}", h.handleDeletePortfolio)
	mux.HandleFunc("GET /api/holdings", h.handleListHoldings)
	mux.HandleFunc("POST /api/holdings", h.handleAddHolding)
	mux.HandleFunc("PUT /api/holdings/{id}", h.handleUpdateHolding)
	mux.HandleFunc("PUT /api/holdings/{id}/reorder", h.handleReorderHolding)
	mux.HandleFunc("DELETE /api/holdings/{id}", h.handleDeleteHolding)
	mux.HandleFunc("GET /api/quote", h.handleQuote)
	mux.HandleFunc("GET /api/chart", h.handleChart)
	mux.HandleFunc("GET /api/etf-holdings", h.handleETFHoldings)
	mux.HandleFunc("GET /api/earnings", h.handleEarnings)
	mux.HandleFunc("GET /api/financials", h.handleFinancials)
	mux.HandleFunc("GET /api/analyst-targets", h.handleAnalystTargets)
	mux.HandleFunc("POST /api/simulate", h.handleSimulate)
	mux.HandleFunc("GET /api/search", h.handleSearch)
}
