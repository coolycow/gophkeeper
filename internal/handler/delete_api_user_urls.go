package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeleteAPIUserURLs возвращает обработчик DELETE /api/user/urls — мягкое удаление URL по списку коротких ключей (JSON-массив строк).
func DeleteAPIUserURLs(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что тип контента - application/json
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			logger.Log.Debug("content type is not application/json")
			_ = c.Error(error.CustomError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})
			return
		}

		// Получаем ID пользователя из запроса.
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Читаем тело запроса
		var keys []string
		err = json.NewDecoder(c.Request.Body).Decode(&keys)

		if err != nil {
			logger.Log.Error("Error decoding request", zap.Error(err))
			_ = c.Error(err)
			return
		}

		err = service.DeleteManyURLs(c.Request.Context(), userID, keys)

		if err != nil {
			logger.Log.Error("Error deleting URLs", zap.Error(err))
			_ = c.Error(err)
			return
		}

		c.Status(http.StatusAccepted)
	}
}
