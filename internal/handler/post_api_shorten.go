package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/service"
	"github.com/coolycow/shortener/internal/shortener"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostAPIShortenHandler возвращает обработчик POST /api/shorten — сокращение URL (JSON: {"url": "..."}).
func PostAPIShortenHandler(urlSvc service.URLService, auditNotifier *audit.Notifier) gin.HandlerFunc {
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

		// Читаем тело запроса
		var req model.APIShortenRequest
		dec := json.NewDecoder(c.Request.Body)

		// Если не удалось декодировать тело запроса, отправляем 400 Bad Request
		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Нормализуем URL
		normalized, err := shortener.NormalizeShortenInput(req.URL)
		if err != nil {
			logger.Log.Debug("normalize shorten input", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Создаём короткую ссылку
		shortURL, err := shortener.Shorten(c.Request.Context(), urlSvc, auditNotifier, userID, normalized)
		if err != nil {
			logger.Log.Debug("cannot create short URL", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Формируем ответ
		logger.Log.Debug("create short url", zap.String("url", shortURL))

		resp := model.APIShortenResponse{
			Result: shortURL,
		}

		// Отправляем ответ
		c.JSON(http.StatusCreated, resp)
	}
}
