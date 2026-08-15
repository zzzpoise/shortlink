package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
)

// generateShort 生成一个随机的短码
func generateShort() string {
	b := make([]byte, 6)                    // 6字节的随机数
	if _, err := rand.Read(b); err != nil { // 填充这些随机数
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b) // 编写为URL安全的Base64字符串
}

// Store 存储短码到长URL的映射，并发安全
type Store struct {
	mu   sync.RWMutex
	urls map[string]string
}

// 创建一个新的Store
func NewStore() *Store {
	return &Store{
		urls: make(map[string]string),
	}
}

// 保存短码和长URL的映射
func (s *Store) Save(short, long string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[short] = long

}

// 根据短码获取长URL
func (s *Store) Get(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, ok := s.urls[short]
	return long, ok
}

//实现HTTP处理器

// 处理post/shorten请求
func shortenHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 只允许post方法
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析请求体json
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		// 验证URL非空
		if req.URL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}
		// 生成并保存短码
		short := generateShort()
		store.Save(short, req.URL)

		// 返回json响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"short": short})
	}
}

// 处理GET请求
func redirectHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//从路径中提取短码
		short := r.URL.Path[1:]
		if short == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		//查找长URL
		long, ok := store.Get(short)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		//302重定向
		http.Redirect(w, r, long, http.StatusFound)
	}
}

// 添加MAIN函数和路由
func main() {
	store := NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", shortenHandler(store))
	mux.HandleFunc("/", redirectHandler(store))

	//启动HTTP服务器
	addr := ":8080"
	println("Server listening on", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
