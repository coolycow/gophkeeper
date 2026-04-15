package handler

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name   string
		method string
		dsn    string
		want   want
	}{
		{
			name:   "Valid DSN",
			method: "GET",
			dsn:    "TEST_DATABASE_DSN",
			want: want{
				code: 200,
			},
		},
		{
			name:   "Invalid DSN",
			method: "GET",
			dsn:    "host=invalid port=9999 user=invalid password=shortener dbname=invalid sslmode=disable",
			want: want{
				code: 500,
			},
		},
		{
			name:   "Method not allowed",
			method: "POST",
			dsn:    "AAbbCC2",
			want: want{
				code: 404,
			},
		},
	}

	cfg, _ := config.InitConfig()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.dsn == "TEST_DATABASE_DSN" {
				test.dsn = os.Getenv("TEST_DATABASE_DSN")
			}

			if test.dsn == "" {
				t.Skip("Skipping integration test. TEST_DATABASE_DSN not set")
			}

			cfg.DatabaseDSN = test.dsn

			repo, err := repository.NewPostgresRepository(test.dsn)

			if err != nil {
				if test.want.code == 500 {
					assert.Nil(t, repo)
					return
				}
			}

			srv := service.NewURLService(cfg, repo)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.GET("/ping", PingHandler(srv))

			request := httptest.NewRequest(test.method, "/ping", nil)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			if err = result.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, test.want.code, result.StatusCode)
		})
	}
}
