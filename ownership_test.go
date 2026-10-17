package stocks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	client "github.com/localitas/localitas-go"
)

const testAPIToken = "lt_test_cluster_token"

func TestMain(m *testing.M) {
	os.Setenv("LOCALITAS_API_TOKEN", testAPIToken)
	client.SetIdentityKeyFromAPIToken(testAPIToken)
	os.Exit(m.Run())
}

// identity signs claims as core does for a request it forwards to the app.
func identity(userID, permission string) string {
	raw, _ := json.Marshal(client.TokenInfo{UserID: userID, Email: userID + "@test.local", Permission: permission})
	payload := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte("localitas-internal-token-v1:" + testAPIToken))
	mac := hmac.New(sha256.New, []byte(base64.StdEncoding.EncodeToString(sum[:])))
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func newOwnershipApp(t *testing.T) (*App, *http.ServeMux) {
	t.Helper()
	app := New(client.New("http://127.0.0.1:1"), "/")
	app.Store = newTestStore(t)
	mux := http.NewServeMux()
	app.RegisterRoutes(mux)
	return app, mux
}

func do(mux *http.ServeMux, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// Another user cannot read, change or delete a portfolio or its holdings,
// whatever ID they send; the owner still can.
func TestPortfolios_BelongToTheirUser(t *testing.T) {
	app, mux := newOwnershipApp(t)
	ctx := context.Background()
	p, err := app.Store.CreatePortfolio(ctx, "alice", "alice retirement")
	if err != nil {
		t.Fatal(err)
	}
	hold, err := app.Store.AddHolding(ctx, "alice", p.ID, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	bob := identity("bob", "write")

	if rec := do(mux, "GET", "/api/portfolios", bob, ""); strings.Contains(rec.Body.String(), p.ID) {
		t.Fatalf("bob's list shows alice's portfolio: %s", rec.Body)
	}
	for _, op := range []struct{ method, path, body string }{
		{"GET", "/api/holdings?portfolio_id=" + p.ID, ""},
		{"PUT", "/api/portfolios/" + p.ID, `{"name":"pwned"}`},
		{"POST", "/api/holdings", `{"portfolio_id":"` + p.ID + `","symbol":"GME"}`},
		{"PUT", "/api/holdings/" + hold.ID, `{"allocation_pct":99}`},
		{"PUT", "/api/holdings/" + hold.ID + "/reorder", `{"position":1}`},
		{"DELETE", "/api/holdings/" + hold.ID, ""},
		{"DELETE", "/api/portfolios/" + p.ID, ""},
		{"POST", "/api/refresh", `{"portfolio_id":"` + p.ID + `"}`},
	} {
		if rec := do(mux, op.method, op.path, bob, op.body); rec.Code != http.StatusNotFound {
			t.Errorf("bob %s %s: %d %s, want 404", op.method, op.path, rec.Code, rec.Body)
		}
	}

	ps, _ := app.Store.ListPortfolios(ctx, "alice")
	if len(ps) != 1 || ps[0].Name != "alice retirement" {
		t.Fatalf("alice's portfolios after bob's attempts: %+v", ps)
	}
	hs, err := app.Store.ListHoldings(ctx, "alice", p.ID)
	if err != nil || len(hs) != 1 || hs[0].Symbol != "AAPL" || hs[0].AllocationPct != 0 {
		t.Fatalf("alice's holdings after bob's attempts: %+v %v", hs, err)
	}

	alice := identity("alice", "write")
	if rec := do(mux, "PUT", "/api/holdings/"+hold.ID, alice, `{"allocation_pct":40}`); rec.Code != http.StatusOK {
		t.Fatalf("alice updates her holding: %d %s", rec.Code, rec.Body)
	}
	if rec := do(mux, "DELETE", "/api/portfolios/"+p.ID, alice, ""); rec.Code != http.StatusOK {
		t.Fatalf("alice deletes her portfolio: %d %s", rec.Code, rec.Body)
	}
	if ps, _ := app.Store.ListAllPortfolios(ctx); len(ps) != 0 {
		t.Fatalf("portfolios left after delete: %+v", ps)
	}
}

func TestPortfolios_RequireAVerifiedUser(t *testing.T) {
	_, mux := newOwnershipApp(t)
	forged := base64.StdEncoding.EncodeToString([]byte(`{"user_id":"alice","permission":"admin"}`))
	for _, token := range []string{"", forged, identity("", "write")} {
		if rec := do(mux, "GET", "/api/portfolios", token, ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("list with %q: %d, want 401", token, rec.Code)
		}
		if rec := do(mux, "POST", "/api/refresh", token, `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("refresh with %q: %d, want 401", token, rec.Code)
		}
	}
	if rec := do(mux, "POST", "/api/portfolios", identity("reader", "read"), `{"name":"x"}`); rec.Code != http.StatusForbidden {
		t.Errorf("a read-only user creates a portfolio: %d, want 403", rec.Code)
	}
}

// Core's cron refreshes every portfolio's symbols; a user's refresh with no
// portfolio named covers only their own.
func TestRefresh_SymbolsByCaller(t *testing.T) {
	app, _ := newOwnershipApp(t)
	ctx := context.Background()
	pa, _ := app.Store.CreatePortfolio(ctx, "alice", "a")
	pb, _ := app.Store.CreatePortfolio(ctx, "bob", "b")
	app.Store.AddHolding(ctx, "alice", pa.ID, "AAPL")
	app.Store.AddHolding(ctx, "bob", pb.ID, "MSFT")
	h := &handler{app: app}

	all, err := h.collectRefreshSymbols(ctx, true, "", "", nil)
	if err != nil || len(all) != 2 {
		t.Fatalf("system refresh symbols = %v, %v; want both users'", all, err)
	}
	own, err := h.collectRefreshSymbols(ctx, false, "bob", "", nil)
	if err != nil || len(own) != 1 || own[0] != "MSFT" {
		t.Fatalf("bob's refresh symbols = %v, %v; want only MSFT", own, err)
	}
}
