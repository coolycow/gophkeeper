package handler

import (
	"context"
	"net/http"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// PingHandler возвращает обработчик GET /ping — проверка доступности хранилища.
func PingHandler(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем доступность хранилища
		if err := service.PingRepository(context.Background()); err != nil {
			_ = c.Error(error.CustomError{
				Message:    "Database connection failed",
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Формируем ответ
		c.String(http.StatusOK, "Database connection succeeded")
	}
}
