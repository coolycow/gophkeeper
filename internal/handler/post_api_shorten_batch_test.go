package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAPIShortenBatchHandler(t *testing.T) {
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
		urls            []string
		gzip            bool
		want            want
	}{
		{
			name:        "Create new",
			method:      http.MethodPost,
			urls:        []string{`https://practicum.yandex.ru`},
			contentType: "application/json",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "URL without domain",
			method:      http.MethodPost,
			urls:        []string{`https://practicum`},
			contentType: "application/json",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Content type with charset",
			method:      http.MethodPost,
			urls:        []string{`https://practicum.yandex.ru`},
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "URL with params",
			method:      http.MethodPost,
			urls:        []string{`https://example.com?a=1,2,3&b=true&c=filter[abd]`},
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Cyrillic URL",
			method:      http.MethodPost,
			urls:        []string{`https://пример.рф/путь?параметр=значение`},
			contentType: "application/json; charset=utf-8",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "Invalid method",
			method:      http.MethodGet,
			urls:        []string{`https://example.com`},
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
			urls:        []string{`https://example.com`},
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
			urls:        []string{``},
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
			urls:        []string{`invalidurl.ru`},
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
			urls:            []string{`https://yandex.ru`},
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
			urls:            []string{`https://yandex.ru`},
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
			router.POST("/api/shorten/batch", PostAPIShortenBatchHandler(urlService))

			var request *http.Request

			var jsonStrings []string
			for i, url := range test.urls {
				jsonStrings = append(jsonStrings, fmt.Sprintf(`{"correlation_id":"%s", "original_url":"%s"}`, strconv.Itoa(i), url))
			}
			jsonData := `[` + strings.Join(jsonStrings, ",") + `]`

			if test.gzip {
				router.Use(gzip.Gzip(gzip.DefaultCompression))
				compressedData, err := gzipData([]byte(jsonData))
				require.NoError(t, err)

				request = httptest.NewRequest(test.method, "/api/shorten/batch", bytes.NewReader(compressedData))
			} else {
				request = httptest.NewRequest(test.method, "/api/shorten/batch", strings.NewReader(jsonData))
			}

			request.Header.Set("Content-Type", test.contentType)
			request.Header.Set("Content-Encoding", test.contentEncoding)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			result := w.Result()

			// Проверяем, что код ответа и тип контента соответствуют ожиданиям
			assert.Equal(t, test.want.code, result.StatusCode)

			resultBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			_ = result.Body.Close()

			if result.StatusCode == http.StatusCreated {
				var resp []model.APIShortenBatchResponse
				err = json.Unmarshal(resultBody, &resp)
				require.NoError(t, err)
			} else {
				assert.Equal(t, test.want.body, string(resultBody))
			}

			if result.StatusCode != http.StatusNotFound {
				require.NoError(t, err)
			}
		})
	}
}
