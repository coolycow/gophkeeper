package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// PostHandler возвращает обработчик POST / — сокращение URL из тела запроса (Content-Type: text/plain).
func PostHandler(service service.URLService, auditNotifier *audit.Notifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что тип контента - text/plain
		// Учитываем, что "Content-Type" может содержать и другие значения, например, charset=utf-8
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			_ = c.Error(error.CustomError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(c.Request.Body)

		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Проверяем, что не пришла пустота
		if len(body) == 0 {
			_ = c.Error(error.CustomError{
				Message:    "Empty body",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Извлекаем строку из тела запроса и проверяем, что она не пуста
		trimBody := strings.TrimSpace(string(body))

		if trimBody == "" {
			_ = c.Error(error.CustomError{
				Message:    "Empty URL",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Парсим URL из строки (фактически проверяем, что это действительно URL)
		validURL, err := url.ParseRequestURI(trimBody)

		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    "Invalid URL",
				StatusCode: http.StatusBadRequest,
			})
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

		shortURL, err := service.CreateShortURL(c.Request.Context(), userID, validURL.String())

		if err != nil {
			_ = c.Error(err)
			return
		}

		// Формируем ответ
		c.String(http.StatusCreated, shortURL)

		// Отправляем событие аудита
		event := audit.NewEvent(audit.ActionShorten, userID, validURL.String())
		auditNotifier.Notify(event)
	}
}
