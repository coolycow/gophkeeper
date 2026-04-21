package grpcserver

import "github.com/coolycow/gophkeeper/internal/observer/audit"

// emitAudit отправляет событие во все настроенные приёмники аудита (файл, URL). Ошибки приёмников не влияют на RPC.
func (s *Server) emitAudit(action, userID, secretID, secretVersionID, attachmentID string) {
	if s.auditNotifier == nil {
		return
	}
	s.auditNotifier.Notify(audit.NewEvent(action, userID, secretID, secretVersionID, attachmentID))
}
