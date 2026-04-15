package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	compgzip "compress/gzip"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAPIShortenHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        string
	}

	tests := []struct {
		name            string
		method          string
		contentType     string
		contentEncoding string
		url             string
		gzip            bool
		want            want
	}{
		{
			name:        "Create new",
			method:      http.MethodPost,
			url:         `https://practicum.yandex.ru`,
			contentType: "application/json",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "URL without domain",
			method:      http.MethodPost,
			url:         `https://practicum`,
			contentType: "application/json",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Content type with charset",
			method:      http.MethodPost,
			url:         `https://practicum.yandex.ru`,
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "URL with params",
			method:      http.MethodPost,
			url:         `https://example.com?a=1,2,3&b=true&c=filter[abd]`,
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Cyrillic URL",
			method:      http.MethodPost,
			url:         `https://пример.рф/путь?параметр=значение`,
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Invalid method",
			method:      http.MethodGet,
			url:         `https://example.com`,
			contentType: "application/json",
			want: want{
				code:        404,
				contentType: "text/plain",
				body:        "404 page not found",
			},
		},
		{
			name:        "Invalid content type",
			method:      http.MethodPost,
			url:         `https://example.com`,
			contentType: "text/plain",
			want: want{
				code:        415,
				contentType: "application/json",
				body:        `{"error":"Content type not allowed"}`,
			},
		},
		{
			name:        "Empty URL",
			method:      http.MethodPost,
			url:         ``,
			contentType: "application/json",
			want: want{
				code:        400,
				contentType: "application/json",
				body:        `{"error":"Empty URL"}`,
			},
		},
		{
			name:        "URL without protocol",
			method:      http.MethodPost,
			url:         `invalidurl.ru`,
			contentType: "application/json",
			want: want{
				code:        400,
				contentType: "application/json",
				body:        `{"error":"Invalid URL"}`,
			},
		},
		{
			name:            "Invalid gzip content",
			method:          http.MethodPost,
			url:             `https://yandex.ru`,
			contentType:     "application/json",
			contentEncoding: "gzip",
			want: want{
				code:        400,
				contentType: "application/json",
				body:        `{"error":"Failed to decompress gzip data"}`,
			},
		},
		{
			name:            "Correct gzip content",
			method:          http.MethodPost,
			url:             `https://yandex.ru`,
			contentType:     "application/json",
			contentEncoding: "gzip",
			gzip:            true,
			want: want{
				code:        201,
				contentType: "application/json",
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

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequestGzip())
			router.Use(middleware.OptionalAuthMiddleware(userService))
			router.POST("/api/shorten", PostAPIShortenHandler(urlService, auditNotifier))

			var request *http.Request

			jsonData := `{"url": "` + test.url + `"}`

			if test.gzip {
				router.Use(gzip.Gzip(gzip.DefaultCompression))
				compressedData, err := gzipData([]byte(jsonData))
				require.NoError(t, err)

				request = httptest.NewRequest(test.method, "/api/shorten", bytes.NewReader(compressedData))
			} else {
				request = httptest.NewRequest(test.method, "/api/shorten", strings.NewReader(jsonData))
			}

			request.Header.Set("Content-Type", test.contentType)
			request.Header.Set("Content-Encoding", test.contentEncoding)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			result := w.Result()

			// Проверяем, что код ответа и тип контента соответствуют ожиданиям
			assert.Equal(t, test.want.code, result.StatusCode)

			_ = result.Body.Close()

			resultBody, err := io.ReadAll(result.Body)

			require.NoError(t, err)

			var resp model.APIShortenResponse
			err = json.Unmarshal(resultBody, &resp)

			if result.StatusCode != http.StatusNotFound {
				require.NoError(t, err)
			}

			// Проверяем, что тело ответа соответствует ожиданиям
			if test.want.code == 201 {
				key := strings.TrimPrefix(resp.Result, cfg.BaseURL+"/")

				// Получаем из репозитория оригинальную ссылку по ключу короткой ссылки
				shortURL, _ := repo.GetShortURL(request.Context(), key)

				// Парсим URL из строки, чтобы корректно сравнивать кириллические адреса
				originalParsedURL, _ := url.Parse(shortURL.OriginalURL)
				bodyParsedURL, _ := url.Parse(test.url)

				// Проверяем, что длина репозитория увеличилась на 1
				assert.Equal(t, repo.GetSize(request.Context()), 1)

				// Проверяем, что оригинальная ссылка и запроса и ссылка из репозитория совпадают
				assert.Equal(t, originalParsedURL.String(), bodyParsedURL.String())

				// Проверяем, что ссылка из ответа соответствует схеме
				assert.Equal(t, resp.Result, cfg.BaseURL+"/"+key)
			} else if test.want.body != "" {
				assert.Equal(t, test.want.body, string(resultBody))
			}
		})
	}
}

// Функция для сжатия данных в GZIP
func gzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := compgzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
