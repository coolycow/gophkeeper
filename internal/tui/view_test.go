package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/stretchr/testify/require"
)

// TestNew_ViewContainsTitle проверяет, что View содержит заголовок
func TestNew_ViewContainsTitle(t *testing.T) {
	cfg, err := config.InitConfigClientWithArgs([]string{"--server-address", "127.0.0.1:50051"})
	require.NoError(t, err)
	m, ok := New(cfg).(*Model)
	require.True(t, ok)
	s := m.View()
	require.Contains(t, s, "GophKeeper")
	require.Contains(t, s, "127.0.0.1:50051")
}

// TestModel_Update_windowSize проверяет, что Update обрабатывает сообщение о размере окна
func TestModel_Update_windowSize(t *testing.T) {
	cfg, err := config.InitConfigClientWithArgs(nil)
	require.NoError(t, err)
	m := New(cfg).(*Model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m2 := next.(*Model)
	require.Equal(t, 100, m2.width)
	require.Equal(t, 40, m2.height)
}
