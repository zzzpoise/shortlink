package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortHandler(t *testing.T) {
	store := NewStore()
	//模拟post请求
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url":"https://go.dev"}`))
	w := httptest.NewRecorder()

	shortenHandler(store)(w, req)

	//检查状态码
	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码200,得到 %d", w.Code)
	}

	//检查响应体包含short字段
	if !strings.Contains(w.Body.String(), "short") {
		t.Errorf("响应缺少short字段： %s", w.Body.String())
	}
}

func TestRedirectHandler(t *testing.T) {
	store := NewStore()
	store.Save("abc123", "https://go.dev")

	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	redirectHandler(store)(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("期望状态码302, 得到 %d", w.Code)
	}

	if loc := w.Header().Get("Location"); loc != "https://go.dev" {
		t.Errorf("重新定向地址错误： %s", loc)
	}
}
