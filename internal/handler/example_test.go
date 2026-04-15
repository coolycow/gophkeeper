package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// ExamplePostHandler демонстрирует сокращение URL через POST / (text/plain).
func ExamplePostHandler() {
	cfg, _ := config.InitConfig()
	tmpFile := "example_post.json"
	_ = os.Remove(tmpFile)
	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()
	srv := service.NewURLService(cfg, repo)
	userSvc := service.NewUserService(cfg, repo)
	notifier := audit.NewNotifier("", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.OptionalAuthMiddleware(userSvc))
	r.POST("/", handler.PostHandler(srv, notifier))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/page"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	fmt.Println(resp.StatusCode)
	// Output: 201
}

// ExamplePostAPIShortenHandler демонстрирует сокращение URL через POST /api/shorten (JSON).
func ExamplePostAPIShortenHandler() {
	cfg, _ := config.InitConfig()
	tmpFile := "example_api_shorten.json"
	_ = os.Remove(tmpFile)
	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()
	srv := service.NewURLService(cfg, repo)
	userSvc := service.NewUserService(cfg, repo)
	notifier := audit.NewNotifier("", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.OptionalAuthMiddleware(userSvc))
	r.POST("/api/shorten", handler.PostAPIShortenHandler(srv, notifier))

	body, _ := json.Marshal(model.APIShortenRequest{URL: "https://practicum.yandex.ru"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	var result model.APIShortenResponse
	_ = json.Unmarshal(respBody, &result)
	fmt.Println(resp.StatusCode, result.Result != "")
	// Output: 201 true
}

// ExampleGetHandler демонстрирует редирект по короткой ссылке GET /:key.
func ExampleGetHandler() {
	cfg, _ := config.InitConfig()
	tmpFile := "example_get.json"
	_ = os.Remove(tmpFile)
	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()
	_, _, _ = repo.SaveURL(context.Background(), "user1", "https://yandex.ru", "abc123")
	srv := service.NewURLService(cfg, repo)
	userSvc := service.NewUserService(cfg, repo)
	notifier := audit.NewNotifier("", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.OptionalAuthMiddleware(userSvc))
	r.GET("/:key", handler.GetHandler(srv, notifier))

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	_ = resp.Body.Close()
	location := resp.Header.Get("Location")
	fmt.Println(resp.StatusCode, location)
	// Output: 307 https://yandex.ru
}

// ExamplePingHandler демонстрирует проверку хранилища GET /ping.
func ExamplePingHandler() {
	cfg, _ := config.InitConfig()
	tmpFile := "example_ping.json"
	_ = os.Remove(tmpFile)
	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()
	srv := service.NewURLService(cfg, repo)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.GET("/ping", handler.PingHandler(srv))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Println(resp.StatusCode, string(body))
	// Output: 200 Database connection succeeded
}

// ExamplePostAPIShortenBatchHandler демонстрирует пакетное сокращение POST /api/shorten/batch.
func ExamplePostAPIShortenBatchHandler() {
	cfg, _ := config.InitConfig()
	tmpFile := "example_batch.json"
	_ = os.Remove(tmpFile)
	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()
	srv := service.NewURLService(cfg, repo)
	userSvc := service.NewUserService(cfg, repo)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.OptionalAuthMiddleware(userSvc))
	r.POST("/api/shorten/batch", handler.PostAPIShortenBatchHandler(srv))

	batch := []model.ShortURL{
		{CorrelationID: "1", OriginalURL: "https://a.com"},
		{CorrelationID: "2", OriginalURL: "https://b.com"},
	}
	body, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	var results []model.APIShortenBatchResponse
	_ = json.Unmarshal(respBody, &results)
	fmt.Println(resp.StatusCode, len(results))
	// Output: 201 2
}
