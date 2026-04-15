package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BenchmarkPostHandler измеряет скорость обработки POST / (создание короткой ссылки).
func BenchmarkPostHandler(b *testing.B) {
	cfg, _ := config.InitConfig()
	tmpFile := "bench_" + uuid.New().String() + ".json"
	defer func() { _ = os.Remove(tmpFile) }()

	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()

	auditNotifier := audit.NewNotifier("", "")
	urlService := service.NewURLService(cfg, repo)
	userService := service.NewUserService(cfg, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.OptionalAuthMiddleware(userService))
	router.POST("/", PostHandler(urlService, auditNotifier))

	var n int
	for b.Loop() {
		b.StopTimer()
		body := fmt.Sprintf("https://example.com/%d", n)
		n++
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetHandler измеряет скорость обработки GET /:key (редирект по короткой ссылке).
func BenchmarkGetHandler(b *testing.B) {
	cfg, _ := config.InitConfig()
	tmpFile := "bench_" + uuid.New().String() + ".json"
	defer func() { _ = os.Remove(tmpFile) }()

	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()

	_, _, _ = repo.SaveURL(context.Background(), "user1", "https://example.com/unique", "Abc123")

	auditNotifier := audit.NewNotifier("", "")
	urlService := service.NewURLService(cfg, repo)
	userService := service.NewUserService(cfg, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.OptionalAuthMiddleware(userService))
	router.GET("/:key", GetHandler(urlService, auditNotifier))

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/Abc123", nil)
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}
