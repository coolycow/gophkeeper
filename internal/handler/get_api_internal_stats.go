package handler

import (
	"net/http"

	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

// GetAPIInternalStatsHandler возвращает обработчик GET /api/internal/stats (данные).
// Доступ по сети обеспечивает middleware.TrustedSubnetInternalStats.
func GetAPIInternalStatsHandler(repo repository.URLRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		c.JSON(http.StatusOK, model.APIInternalStats{
			Urls:  repo.GetSize(ctx),
			Users: repo.GetUsersCount(ctx),
		})
	}
}
