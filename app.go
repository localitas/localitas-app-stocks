package stocks

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

var logger = slog.Default().With("component", "stocks")

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
			logger.Warn("install attempt failed, retrying", "attempt", attempt, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			logger.Warn("install migrations attempt failed, retrying", "attempt", attempt, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		logger.Error("index template error", "error", err)
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
	// Portfolios belong to the user core vouches for; changing one needs
	// write scope. Market data needs only a verified user.
	user := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, client.AuthMiddleware(handler))
	}
	writer := func(pattern string, handler http.HandlerFunc) {
		user(pattern, client.RequireScopeFunc(client.ScopeWrite, handler))
	}
	user("GET /api/portfolios", h.handleListPortfolios)
	writer("POST /api/portfolios", h.handleCreatePortfolio)
	writer("PUT /api/portfolios/{id}", h.handleUpdatePortfolio)
	writer("DELETE /api/portfolios/{id}", h.handleDeletePortfolio)
	user("GET /api/holdings", h.handleListHoldings)
	writer("POST /api/holdings", h.handleAddHolding)
	writer("PUT /api/holdings/{id}", h.handleUpdateHolding)
	writer("PUT /api/holdings/{id}/reorder", h.handleReorderHolding)
	writer("DELETE /api/holdings/{id}", h.handleDeleteHolding)
	user("GET /api/quote", h.handleQuote)
	user("GET /api/chart", h.handleChart)
	user("GET /api/etf-holdings", h.handleETFHoldings)
	user("GET /api/earnings", h.handleEarnings)
	user("GET /api/financials", h.handleFinancials)
	user("GET /api/analyst-targets", h.handleAnalystTargets)
	user("POST /api/simulate", h.handleSimulate)
	user("GET /api/search", h.handleSearch)
	// Refresh serves core's cron (every portfolio) and users (their own).
	refreshAsUser := client.AuthMiddleware(http.HandlerFunc(h.handleRefresh))
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		if client.IsSystemRequest(r) {
			h.handleRefresh(w, r)
			return
		}
		refreshAsUser.ServeHTTP(w, r)
	})
}
