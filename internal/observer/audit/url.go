package audit

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/hashicorp/go-retryablehttp"
)

// defaultRetryClient — HTTP-клиент с ретраями для отправки аудита.
var defaultRetryClient = retryablehttp.NewClient()

// URLReceiver отправляет события аудита на удалённый сервер методом POST.
type URLReceiver struct {
	url    string
	client *retryablehttp.Client
}

// NewURLReceiver создаёт приёмник на удалённый URL.
func NewURLReceiver(auditURL string) *URLReceiver {
	return &URLReceiver{url: auditURL, client: defaultRetryClient}
}

// Send отправляет событие POST-запросом с JSON-телом (с автоматическими ретраями).
func (u *URLReceiver) Send(event *model.Audit) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, u.url, data)
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("send audit event to %s: %w", u.url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit endpoint returned status %d", resp.StatusCode)
	}
	return nil
}
