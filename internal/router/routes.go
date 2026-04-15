package router

import (
	"net"
	"strings"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// parseTrustedSubnet парсит строку CIDR в структуру net.IPNet
func parseTrustedSubnet(cidr string) (*net.IPNet, error) {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	return ipNet, err
}

// setupURLRoutes настраивает маршруты для URL-сервиса
func setupURLRoutes(
	r *gin.Engine,
	cfg *config.Config,
	repo repository.URLRepository,
	auditNotifier *audit.Notifier,
) {
	srv := service.NewURLService(cfg, repo)
	cookieService := service.NewUserService(cfg, repo)

	// Парсим trusted subnet
	trusted, err := parseTrustedSubnet(cfg.TrustedSubnet)
	if err != nil {
		trusted = nil
	}

	r.GET("/ping", handler.PingHandler(srv))

	// Маршрут для получения статистики
	// Использует middleware.TrustedSubnetInternalStats для ограничения доступа по trusted subnet
	r.GET("/api/internal/stats", middleware.TrustedSubnetInternalStats(trusted), handler.GetAPIInternalStatsHandler(repo))

	// Группа маршрутов с опциональной аутентификацией
	authGroup := r.Group("/")
	authGroup.Use(middleware.OptionalAuthMiddleware(cookieService))

	authGroup.POST("/", handler.PostHandler(srv, auditNotifier))
	authGroup.GET("/:key", handler.GetHandler(srv, auditNotifier))

	authGroup.POST("/api/shorten", handler.PostAPIShortenHandler(srv, auditNotifier))
	authGroup.POST("/api/shorten/batch", handler.PostAPIShortenBatchHandler(srv))

	authGroup.GET("/api/user/urls", handler.GetAPIUserURLs(srv))
	authGroup.DELETE("/api/user/urls", handler.DeleteAPIUserURLs(srv))
}
