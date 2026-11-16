---
title: Stocks
description: Stock portfolio tracking and market data
---

# Stocks

Track stock portfolios, view real-time quotes, and analyze financial data.

## Portfolios

Create portfolios to organize your stock holdings. Each portfolio has a name and contains a list of holdings.

**GET /api/portfolios** - List all portfolios
**POST /api/portfolios** - Create a new portfolio
**PUT /api/portfolios/{id}** - Update portfolio name
**DELETE /api/portfolios/{id}** - Delete a portfolio

## Holdings

Add stock ticker symbols to a portfolio with quantity and cost basis.

**GET /api/holdings** - List holdings in a portfolio
**POST /api/holdings** - Add a holding to a portfolio
**PUT /api/holdings/{id}** - Update holding details (shares, cost)
**PUT /api/holdings/{id}/reorder** - Change display order of a holding
**DELETE /api/holdings/{id}** - Remove a holding

## Market Data

Fetch live stock data from Yahoo Finance.

**GET /api/quote?symbol={ticker}** - Get a real-time quote for a ticker
**GET /api/chart?symbol={ticker}** - Get historical price chart data
**GET /api/etf-holdings?symbol={ticker}** - View top holdings of an ETF
**GET /api/earnings?symbol={ticker}** - View earnings history and estimates
**GET /api/financials?symbol={ticker}** - View income statement and balance sheet
**GET /api/analyst-targets?symbol={ticker}** - View analyst price targets

## Simulation

**POST /api/simulate** - Run a portfolio simulation with hypothetical trades

## Search

**GET /api/search?q={query}** - Search for stock tickers by name or symbol

## Build & Deploy

Run these in the app's repository, `~/localitas-app-stocks`.

### Version

```bash
bin/stocks-server --version
```

### Build from source

```bash
# Native binary (bin/stocks-server)
make build

# Linux amd64 binary the Docker image is built from (stocks-server-linux-amd64)
make build-linux
```

Both run `make lint` (gofmt and go vet) first; `make test` runs the lint step and then the tests.

### Run natively

```bash
make start   # builds bin/stocks-server, writes the launchd service com.localitas.app.stocks for bin/stocks-server and loads it on port 9210
make stop
make logs    # follows ~/.localitas/logs/stocks/stdout.log
```

The app announces itself to core when it starts, and core serves it at `/apps/ext/stocks/`.

### Docker

```bash
make build-docker   # builds the Linux binary and the stocks:latest image
make start-docker   # runs container "stocks" on port 9210, labelled localitas.app=true
make stop-docker
make logs-docker
```

Core registers containers labelled `localitas.app=true` once Docker reports them healthy.

### Publish

```bash
# Run the tests, build the image, push ghcr.io/localitas/localitas-app-stocks:latest and :<VERSION>
make docker-push

# Build dist/stocks-server-darwin-arm64 and dist/stocks-server-darwin-amd64
# and create GitHub release vX.Y.Z on this repository with them
make release VERSION=vX.Y.Z
```

`VERSION` defaults to `git describe --tags --always --dirty`.

### Download

Pre-built macOS binaries are on the [releases page](https://github.com/localitas/localitas-app-stocks/releases):
- `stocks-server-darwin-arm64` (Apple Silicon)
- `stocks-server-darwin-amd64` (Intel)

Every push to `main` publishes a release: `.github/workflows/release.yml` runs the tests and releases both binaries as `v0.1.<commit count>.<UTC timestamp>`. `make release` publishes one by hand.

Download with the GitHub CLI:

    gh release download --repo localitas/localitas-app-stocks --pattern 'stocks-server-*'
