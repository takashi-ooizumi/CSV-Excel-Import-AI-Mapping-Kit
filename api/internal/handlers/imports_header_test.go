package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLooksLikeHeader(t *testing.T) {
	// ヘッダあり
	first := []string{"order_id", "customer_id", "unit_price"}
	next := []string{"1001", "1", "980"}
	if !looksLikeHeader(first, next) { // ← 2引数で呼ぶ
		t.Fatalf("expected header=true, got false")
	}

	// ログ形式（ヘッダなし）
	first = []string{"2024-07-01 10:01:23", "INFO", "login", "user=1", "ip=203.0.113.10"}
	next = []string{"2024-07-01 10:05:10", "WARN", "retry", "user=1", "count=2"}
	if looksLikeHeader(first, next) {
		t.Fatalf("expected header=false, got true")
	}
}

func TestCORSMiddleware(t *testing.T) {
	h := CORSMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("OPTIONS", "/api/imports", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("missing/invalid ACAO header: %q", got)
	}
}

func TestNormalizeHeader(t *testing.T) {
	in := []string{"order_id", "order_id", "Unit Price"}
	out := normalizeHeaders(in)
	got := strings.Join(out, ",")
	want := "order_id,order_id_1,unit_price"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
