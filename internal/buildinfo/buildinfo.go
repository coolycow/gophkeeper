// Package buildinfo предоставляет информацию о сборке приложения.
// Используется как для сервера, так и для клиента.
package buildinfo

import (
	"fmt"
	"io"
	"strings"
)

// Version версия приложения (ldflags: -X main.buildVersion=...).
var Version string

// BuildDate дата сборки (ldflags: -X main.buildDate=...).
var BuildDate string

// BuildCommit коммит (ldflags: -X main.buildCommit=...).
var BuildCommit string

// OrNA возвращает s, если не пусто, иначе "N/A".
func OrNA(s string) string {
	if strings.TrimSpace(s) == "" {
		return "N/A"
	}
	return s
}

// Fprint выводит версию, дату сборки и коммит в w, каждый на своей строке.
func Fprint(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Build version: %s\n", OrNA(Version))
	_, _ = fmt.Fprintf(w, "Build date: %s\n", OrNA(BuildDate))
	_, _ = fmt.Fprintf(w, "Build commit: %s\n", OrNA(BuildCommit))
}
