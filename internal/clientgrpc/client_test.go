package clientgrpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNeedsRefreshSoon(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	exp := now.Add(10 * time.Second)
	assert.True(t, NeedsRefreshSoon(exp, 30*time.Second, now))
	assert.False(t, NeedsRefreshSoon(exp, 5*time.Second, now))
	assert.False(t, NeedsRefreshSoon(time.Time{}, time.Minute, now))
}
