package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAPIInternalStatsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, trustedClassA, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	tests := []struct {
		name       string
		trusted    *net.IPNet
		realIP     string
		wantCode   int
		wantParsed *model.APIInternalStats
	}{
		{
			name:     "empty trusted subnet denies",
			trusted:  nil,
			realIP:   "127.0.0.1",
			wantCode: http.StatusForbidden,
		},
		{
			name:     "missing X-Real-IP",
			trusted:  trustedClassA,
			wantCode: http.StatusForbidden,
		},
		{
			name:     "IP outside subnet",
			trusted:  trustedClassA,
			realIP:   "192.168.1.1",
			wantCode: http.StatusForbidden,
		},
		{
			name:     "invalid X-Real-IP",
			trusted:  trustedClassA,
			realIP:   "not-an-ip",
			wantCode: http.StatusForbidden,
		},
		{
			name:     "allowed IP",
			trusted:  trustedClassA,
			realIP:   "10.1.2.3",
			wantCode: http.StatusOK,
			wantParsed: &model.APIInternalStats{
				Urls:  2,
				Users: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := "test_stats_" + uuid.New().String() + ".json"
			defer func() { _ = os.Remove(tmpFile) }()

			repo := repository.NewDoubleMapsRepository(tmpFile)
			defer func() { _ = repo.Close() }()

			ctx := context.Background()
			_, _, errSave := repo.SaveURL(ctx, "u1", "https://a.example", "keyaaaaa")
			require.NoError(t, errSave)
			_, _, errSave = repo.SaveURL(ctx, "u1", "https://b.example", "keybbbbbb")
			require.NoError(t, errSave)
			_, errUser := repo.CreateUser(ctx)
			require.NoError(t, errUser)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.GET("/api/internal/stats", middleware.TrustedSubnetInternalStats(tt.trusted), GetAPIInternalStatsHandler(repo))

			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			assert.Equal(t, tt.wantCode, res.StatusCode)

			if tt.wantParsed != nil {
				var got model.APIInternalStats
				require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
				assert.Equal(t, *tt.wantParsed, got)
			}
		})
	}
}
