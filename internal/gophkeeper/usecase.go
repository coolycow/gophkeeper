package shortener

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	httperr "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/service"
)

// NormalizeShortenInput проверяет непустоту и корректность URL (как POST /api/shorten).
func NormalizeShortenInput(raw string) (string, error) {
	trimURL := strings.TrimSpace(raw)
	if trimURL == "" {
		return "", httperr.CustomError{
			Message:    "Empty URL",
			StatusCode: http.StatusBadRequest,
		}
	}

	validURL, err := url.ParseRequestURI(trimURL)
	if err != nil {
		return "", httperr.CustomError{
			Message:    "Invalid URL",
			StatusCode: http.StatusBadRequest,
		}
	}

	return validURL.String(), nil
}

// Shorten создаёт короткую ссылку и публикует событие аудита (как успешный POST /api/shorten).
func Shorten(ctx context.Context, urlSvc service.URLService, auditNotifier *audit.Notifier, userID, normalizedURL string) (shortURL string, err error) {
	shortURL, err = urlSvc.CreateShortURL(ctx, userID, normalizedURL)
	if err != nil {
		return "", err
	}

	auditNotifier.Notify(audit.NewEvent(audit.ActionShorten, userID, normalizedURL))
	return shortURL, nil
}

// Expand возвращает исходный URL; deleted == true соответствует HTTP 410 Gone.
func Expand(ctx context.Context, urlSvc service.URLService, auditNotifier *audit.Notifier, userID, key string) (originalURL string, deleted bool, err error) {
	key = strings.TrimSpace(strings.TrimPrefix(key, "/"))
	su, err := urlSvc.GetShortURL(ctx, key)
	if err != nil {
		return "", false, err
	}

	if su.DeletedAt != nil {
		return "", true, nil
	}

	auditNotifier.Notify(audit.NewEvent(audit.ActionFollow, userID, su.OriginalURL))
	return su.OriginalURL, false, nil
}

// ListURLs формирует список ссылок пользователя в том же виде, что JSON GET /api/user/urls.
func ListURLs(ctx context.Context, urlSvc service.URLService, userID string) ([]model.APIUserURLsResponse, error) {
	shortURLs, err := urlSvc.GetManyShortURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(shortURLs) == 0 {
		return nil, nil
	}

	base := urlSvc.GetBaseURL()
	out := make([]model.APIUserURLsResponse, 0, len(shortURLs))
	for _, u := range shortURLs {
		out = append(out, u.ToAPIUserURLsResponse(base))
	}

	return out, nil
}
