package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostAPIShortenBatchHandler возвращает обработчик POST /api/shorten/batch — пакетное сокращение URL (JSON-массив с correlation_id, original_url).
func PostAPIShortenBatchHandler(service service.URLService) gin.HandlerFunc {
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

		var URLs []model.ShortURL
		err := json.NewDecoder(c.Request.Body).Decode(&URLs)

		if err != nil {
			logger.Log.Error("Error unmarshalling request body", zap.Error(err))
			_ = c.Error(err)
			return
		}

		if len(URLs) == 0 {
			_ = c.Error(error.CustomError{
				Message:    "Empty array of URLs",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		err = service.CreateManyShortURL(c.Request.Context(), userID, URLs)

		if err != nil {
			logger.Log.Debug("cannot create short URLs", zap.Error(err))
			_ = c.Error(err)
			return
		}

		responses := make([]model.APIShortenBatchResponse, 0, len(URLs))

		for _, u := range URLs {
			responses = append(responses, u.ToAPIShortenBatchResponse(service.GetBaseURL()))
		}

		c.JSON(http.StatusCreated, responses)
	}
}
