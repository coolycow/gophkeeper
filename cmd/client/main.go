// Клиент GophKeeper: TUI на Bubble Tea и обмен с сервером по gRPC.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/coolycow/gophkeeper/internal/buildinfo"
	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/tui"
)

// Сюда подставляются значения при сборке через -ldflags (см. README.md).
var buildVersion, buildDate, buildCommit string

// init инициализирует информацию о сборке
func init() {
	buildinfo.Version = buildVersion    // ldflags: -X main.buildVersion=...
	buildinfo.BuildDate = buildDate     // ldflags: -X main.buildDate=...
	buildinfo.BuildCommit = buildCommit // ldflags: -X main.buildCommit=...
}

func main() {
	// Инициализируем конфигурацию клиента
	cfg, err := config.InitConfigClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "конфигурация: %v\n", err)
		os.Exit(1)
	}

	// Если нужно показать версию, выводим её и завершаем работу
	if cfg.ShowVersion {
		buildinfo.Fprint(os.Stdout)
		return
	}

	// Инициализируем программу
	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen())

	// Запускаем программу
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка клиента: %v\n", err)
		os.Exit(1)
	}
}
