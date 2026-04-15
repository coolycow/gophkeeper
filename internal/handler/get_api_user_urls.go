package handler

import (
	"net/http"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/service"
	"github.com/coolycow/shortener/internal/shortener"
	"github.com/gin-gonic/gin"
)

// GetAPIUserURLs возвращает обработчик GET /api/user/urls — список всех сокращённых URL текущего пользователя.
func GetAPIUserURLs(urlSvc service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// Получаем список ссылок для текущего пользователя
		responses, err := shortener.ListURLs(c.Request.Context(), urlSvc, userID)
		if err != nil {
			_ = c.Error(err)
			return
		}

		// Если список ссылок пуст, отправляем 204 No Content
		if len(responses) == 0 {
			c.JSON(http.StatusNoContent, nil)
			return
		}

		// Отправляем список ссылок в формате JSON
		c.JSON(http.StatusOK, responses)
	}
}
