package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        string // Добавляем поле для проверки тела ответа
	}

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		want        want
	}{
		{
			name:        "Create new",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "text/plain",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "URL without domain",
			method:      http.MethodPost,
			body:        "https://example",
			contentType: "text/plain",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Content type with charset",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "URL with params",
			method:      http.MethodPost,
			body:        "https://example.com?a=1,2,3&b=true&c=filter[abd]",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Cyrillic URL",
			method:      http.MethodPost,
			body:        "https://пример.рф/путь?параметр=значение",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Duplicate URL",
			method:      http.MethodPost,
			body:        "https://duplicate-example.com",
			contentType: "text/plain",
			want: want{
				code:        409,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Invalid method",
			method:      http.MethodGet,
			body:        "https://example.com",
			contentType: "text/plain",
			want: want{
				code:        404,
				contentType: "text/plain",
				body:        "404 page not found",
			},
		},
		{
			name:        "Invalid content type",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "application/json",
			want: want{
				code:        415,
				contentType: "text/plain",
				body:        "Content type not allowed",
			},
		},
		{
			name:        "Empty body",
			method:      http.MethodPost,
			body:        "",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Empty body",
			},
		},
		{
			name:        "Empty URL",
			method:      http.MethodPost,
			body:        "           ",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Empty URL",
			},
		},
		{
			name:        "URL without protocol",
			method:      http.MethodPost,
			body:        "invalidurl.ru",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Invalid URL",
			},
		},
	}

	cfg, _ := config.InitConfig()

	auditNotifier := audit.NewNotifier("", "")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Временный файл для хранилища
			tmpFile := "test_" + uuid.New().String() + ".json"
			defer func() { _ = os.Remove(tmpFile) }()

			repo := repository.NewDoubleMapsRepository(tmpFile)
			defer func() { _ = repo.Close() }()

			urlService := service.NewURLService(cfg, repo)
			userService := service.NewUserService(cfg, repo)

			if test.name == "Duplicate URL" {
				_, _, _ = repo.SaveURL(context.Background(), uuid.New().String(), test.body, "Dup123456")
			}

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.OptionalAuthMiddleware(userService))
			router.POST("/", PostHandler(urlService, auditNotifier))

			request := httptest.NewRequest(test.method, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			result := w.Result()

			// Проверяем, что код ответа и тип контента соответствуют ожиданиям
			assert.Equal(t, test.want.code, result.StatusCode)

			_ = result.Body.Close()

			resultBody, err := io.ReadAll(result.Body)

			resultString := string(resultBody)

			require.NoError(t, err)

			// Проверяем, что тело ответа соответствует ожиданиям
			if test.want.code == 201 {
				key := strings.TrimPrefix(resultString, cfg.BaseURL+"/")

				// Получаем из репозитория оригинальную ссылку по ключу короткой ссылки
				shortURL, _ := repo.GetShortURL(request.Context(), key)

				// Парсим URL из строки, чтобы корректно сравнивать кириллические адреса
				originalParsedURL, _ := url.Parse(shortURL.OriginalURL)
				bodyParsedURL, _ := url.Parse(test.body)

				// Проверяем, что длина репозитория увеличилась на 1
				assert.Equal(t, repo.GetSize(request.Context()), 1)

				// Проверяем, что оригинальная ссылка и запроса и ссылка из репозитория совпадают
				assert.Equal(t, originalParsedURL.String(), bodyParsedURL.String())

				// Проверяем, что ссылка из ответа соответствует схеме
				assert.Equal(t, resultString, cfg.BaseURL+"/"+key)
			} else if test.want.body != "" {
				assert.Equal(t, test.want.body, resultString)
			}
		})
	}
}
