package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coolycow/shortener/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAPIUserURLs(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "Empty repository",
			method: "GET",
			want: want{
				code: 204,
			},
		},
		{
			name:   "Filled repository",
			method: "GET",
			want: want{
				code: 200,
			},
		},
		{
			name:   "Method not allowed",
			method: "POST",
			want: want{
				code: 404,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var addDefault bool

			if test.name != "Empty repository" {
				addDefault = true
			}

			urlService, userService := setupTestService(addDefault)

			user, err := userService.CreateUser(context.Background())
			require.NoError(t, err)

			cookieValue, err := userService.GetCookieValueByUserID(user.ID)
			require.NoError(t, err)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequiredAuthMiddleware(userService))
			router.GET("/api/user/urls", GetAPIUserURLs(urlService))

			request := httptest.NewRequest(test.method, "/api/user/urls", nil)

			request.AddCookie(&http.Cookie{
				Name:  "auth",
				Value: cookieValue,
			})

			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			if err := result.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, test.want.code, result.StatusCode)
		})
	}
}
