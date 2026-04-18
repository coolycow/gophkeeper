// Package grpcserver поднимает gRPC GophKeeperService с тем же TLS, что и HTTP (EnableHTTPS + сертификаты).
package grpcserver

import (
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/coolycow/gophkeeper/internal/service"
	"google.golang.org/grpc"
)

// Server реализует gophkeeperpb.GophKeeperServiceServer, вызывая пакет gophkeeper.
type Server struct {
	gophkeeperpb.UnimplementedGophKeeperServiceServer
	auditNotifier    *audit.Notifier
	userSvc          service.UserService
	secretSvc        service.SecretService
	secretVersionSvc service.SecretVersionService
	attachmentSvc    service.AttachmentService
}

// NewServer собирает gRPC-обработчик с зависимостями, совпадающими с HTTP-хендлерами.
func NewServer(auditNotifier *audit.Notifier,
	userSvc service.UserService, secretSvc service.SecretService,
	secretVersionSvc service.SecretVersionService, attachmentSvc service.AttachmentService) *Server {
	return &Server{
		auditNotifier:    auditNotifier,
		userSvc:          userSvc,
		secretSvc:        secretSvc,
		secretVersionSvc: secretVersionSvc,
		attachmentSvc:    attachmentSvc,
	}
}

// Register регистрирует сервис на переданном grpc.Server.
func (s *Server) Register(reg grpc.ServiceRegistrar) {
	gophkeeperpb.RegisterGophKeeperServiceServer(reg, s)
}
