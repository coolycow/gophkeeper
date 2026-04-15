// Package handler содержит HTTP-обработчики сервиса коротких ссылок.
package handler

import (
	"net/http"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/service"
	"github.com/coolycow/shortener/internal/shortener"
	"github.com/gin-gonic/gin"
)

// GetHandler возвращает обработчик GET /:key — редирект по короткой ссылке на оригинальный URL.
func GetHandler(urlSvc service.URLService, auditNotifier *audit.Notifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")

		// Получаем ID пользователя из запроса
		userID, err := middleware.GetUserIDFromGinContext(c)

		// Если не удалось получить ID пользователя, отправляем 500 Internal Server Error
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Разворачиваем короткую ссылку
		orig, deleted, err := shortener.Expand(c.Request.Context(), urlSvc, auditNotifier, userID, key)
		if err != nil {
			_ = c.Error(err)
			return
		}

		// Если ссылка удалена, отправляем 410 Gone
		if deleted {
			c.Status(http.StatusGone)
			return
		}

		// Устанавливаем заголовок Location и статус 307 Temporary Redirect
		c.Header("Location", orig)
		c.Status(http.StatusTemporaryRedirect)
	}
}
