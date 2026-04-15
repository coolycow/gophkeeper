package handler

import (
	"context"
	"net/http/httptest"
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
)

// setupFullRepository возвращает репозиторий с заранее сохраненными ключами и соответствующими им URL
func setupTestService(addDefaultURLs bool) (service.URLService, service.UserService) {
	cfg, _ := config.InitConfig()

	// Временный файл для хранилища
	tmpFile := "test_" + uuid.New().String() + ".json"
	defer func() { _ = os.Remove(tmpFile) }()

	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer func() { _ = repo.Close() }()

	if addDefaultURLs {
		defaultURLs := map[string]string{
			"":          "https://mail.ru",
			"AAbbCC1":   "https://google.com",
			"AAbbCC2":   "https://rambler.com",
			"AAbbCC3":   "https://yandex.com",
			"AAbbCC3$#": "https:/vk.com",
			"Dup123456": "https://duplicate-example.com",
		}

		for key, u := range defaultURLs {
			_, _, _ = repo.SaveURL(context.Background(), uuid.New().String(), u, key)
		}
	}

	return service.NewURLService(cfg, repo), service.NewUserService(cfg, repo)
}

func TestGetHandler(t *testing.T) {
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name   string
		method string
		key    string
		want   want
	}{
		{
			name:   "Valid URL",
			method: "GET",
			key:    "AAbbCC1",
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "ID with params",
			method: "GET",
			key:    "AAbbCC2?param=value",
			want: want{
				code:     307,
				location: "https://rambler.com",
			},
		},
		{
			name:   "ID with special chars",
			method: "GET",
			key:    "AAbbCC3$#",
			want: want{
				code:     307,
				location: "https:/vk.com",
			},
		},
		{
			name:   "Start spaces",
			method: "GET",
			key:    "%20%20%20%20AAbbCC1",
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "End spaces",
			method: "GET",
			key:    "AAbbCC2%20%20%20%20",
			want: want{
				code:     307,
				location: "https://rambler.com",
			},
		},
		{
			name:   "Method not allowed",
			method: "POST",
			key:    "AAbbCC2",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Empty ID",
			method: "GET",
			key:    "",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Not found ID",
			method: "GET",
			key:    "AAbbCC4",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Incorrect URL",
			method: "GET",
			key:    "AAbbCC4/extrapath",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "ID with spaces",
			method: "GET",
			key:    "AAbbCC4%20extrapath1%20extrapath2",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Long ID",
			method: "GET",
			key:    strings.Repeat("C", 1000),
			want: want{
				code:     404,
				location: "",
			},
		},
	}

	auditNotifier := audit.NewNotifier("", "")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService, userService := setupTestService(true)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.OptionalAuthMiddleware(userService))
			router.GET("/:key", GetHandler(urlService, auditNotifier))

			request := httptest.NewRequest(test.method, "/"+test.key, nil)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			if err := result.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, test.want.code, result.StatusCode)
			assert.Equal(t, test.want.location, result.Header.Get("Location"))
		})
	}
}
