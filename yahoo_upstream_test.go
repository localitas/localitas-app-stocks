package stocks

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// yahooStandIn stands in for Yahoo Finance: /fc sets the session cookie,
// /crumb hands the crumb to a holder of it, and /data and /plain fail their
// first request with a 503, then answer; /data only to the cookie and crumb.
func yahooStandIn(t *testing.T) (*httptest.Server, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var dataCalls, plainCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /fc", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A1", Value: "session"})
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /crumb", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("A1"); err != nil || c.Value != "session" {
			http.Error(w, "no session", http.StatusForbidden)
			return
		}
		w.Write([]byte("crumb1"))
	})
	mux.HandleFunc("GET /data", func(w http.ResponseWriter, r *http.Request) {
		if dataCalls.Add(1) == 1 {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		if c, err := r.Cookie("A1"); err != nil || c.Value != "session" || r.URL.Query().Get("crumb") != "crumb1" || r.UserAgent() != "Mozilla/5.0" {
			http.Error(w, "no session", http.StatusForbidden)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("GET /plain", func(w http.ResponseWriter, r *http.Request) {
		if plainCalls.Add(1) == 1 {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"plain":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	oldCookieURL, oldCrumbURL, oldCrumb, oldCookies := yahooCookieURL, yahooCrumbURL, yahooCrumb, yahooCookies
	yahooCookieURL, yahooCrumbURL, yahooCrumb, yahooCookies = srv.URL+"/fc", srv.URL+"/crumb", "", nil
	t.Cleanup(func() {
		yahooCookieURL, yahooCrumbURL, yahooCrumb, yahooCookies = oldCookieURL, oldCrumbURL, oldCrumb, oldCookies
	})
	return srv, &dataCalls, &plainCalls
}

// Yahoo failing for a moment is asked again, and a request that needs the
// session still carries its cookie and crumb.
func TestYahoo_AMomentaryFailureIsAskedAgain(t *testing.T) {
	srv, dataCalls, plainCalls := yahooStandIn(t)

	body, err := yahooAuthGet(srv.URL + "/data?modules=financialData")
	if err != nil || string(body) != `{"ok":true}` || dataCalls.Load() != 2 {
		t.Errorf("session request = %q, %v after %d calls, want the answer after 2", body, err, dataCalls.Load())
	}
	body, err = yahooGet(srv.URL + "/plain")
	if err != nil || string(body) != `{"plain":true}` || plainCalls.Load() != 2 {
		t.Errorf("plain request = %q, %v after %d calls, want the answer after 2", body, err, plainCalls.Load())
	}
}
