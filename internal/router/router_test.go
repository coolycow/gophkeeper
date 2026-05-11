package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Маршрут /debug/pprof/ открыт после инициализации роутера.
func TestNewRouter_PprofIndex(t *testing.T) {
	cfg := &config.ConfigServer{}
	n, err := audit.NewNotifier("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = n.Close() })

	r := NewRouter(cfg, nil, n)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
