package router

import (
	"net"
	"strings"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/repository"
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
	cfg *config.ConfigServer,
	repo repository.GophKeeperRepository,
	auditNotifier *audit.Notifier) {
}
