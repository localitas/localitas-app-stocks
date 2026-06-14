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

### Version

```bash
./stocks-server --version
```

### Build from source

```bash
# Development (native)
cd apps/stocks && go build -o bin/stocks-server ./cmd/stocks-server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o bin/stocks-server-linux-amd64 ./cmd/stocks-server
```

### Docker

Build a Docker image directly from the binary:

```bash
# Default base image (debian:12-slim)
./stocks-server docker-build

# Custom base image
./stocks-server docker-build --base ubuntu:24.04

# Custom Dockerfile
./stocks-server docker-build --dockerfile ./my.Dockerfile

# Tag and push to registry
./stocks-server docker-build --tag ghcr.io/localitas/stocks:latest --push
```

The `docker-build` command requires a Linux amd64 binary in the same directory. Run `make deploy-build` from the project root first.

### Download

Pre-built binaries are available on the [GitHub releases page](https://github.com/localitas/localitas/releases).

Each release includes three builds per app:
- `stocks-server-darwin-arm64` (macOS Apple Silicon)
- `stocks-server-linux-amd64` (Linux x86_64)
- `stocks-server-linux-arm64` (Linux ARM64)

Download with the GitHub CLI:

    gh release download --repo localitas/localitas --pattern 'stocks-server-*'

### Release

All app binaries are published to GitHub releases as part of `make deploy-upload-image`.
