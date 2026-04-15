// Package grpcserver поднимает gRPC ShortenerService с тем же TLS, что и HTTP (EnableHTTPS + сертификаты).
package grpcserver

import (
	"context"

	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/proto/shortenerpb"
	"github.com/coolycow/shortener/internal/service"
	"github.com/coolycow/shortener/internal/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует shortenerpb.ShortenerServiceServer, вызывая пакет shortener.
type Server struct {
	shortenerpb.UnimplementedShortenerServiceServer
	url           service.URLService
	auditNotifier *audit.Notifier
}

// NewServer собирает gRPC-обработчик с зависимостями, совпадающими с HTTP-хендлерами.
func NewServer(urlSvc service.URLService, auditNotifier *audit.Notifier) *Server {
	return &Server{
		url:           urlSvc,
		auditNotifier: auditNotifier,
	}
}

// Register регистрирует сервис на переданном grpc.Server.
func (s *Server) Register(reg grpc.ServiceRegistrar) {
	shortenerpb.RegisterShortenerServiceServer(reg, s)
}

// ShortenURL — аналог POST /api/shorten (JSON result).
func (s *Server) ShortenURL(ctx context.Context, req *shortenerpb.URLShortenRequest) (*shortenerpb.URLShortenResponse, error) {
	userID, ok := shortener.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "user id missing in context")
	}

	normalized, err := shortener.NormalizeShortenInput(req.GetUrl())
	if err != nil {
		return nil, grpcError(err)
	}

	result, err := shortener.Shorten(ctx, s.url, s.auditNotifier, userID, normalized)
	if err != nil {
		return nil, grpcError(err)
	}

	return shortenerpb.URLShortenResponse_builder{Result: result}.Build(), nil
}

// ExpandURL — аналог GET /:id; при удалённой ссылке — FailedPrecondition (как 410 Gone).
func (s *Server) ExpandURL(ctx context.Context, req *shortenerpb.URLExpandRequest) (*shortenerpb.URLExpandResponse, error) {
	userID, ok := shortener.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "user id missing in context")
	}

	orig, deleted, err := shortener.Expand(ctx, s.url, s.auditNotifier, userID, req.GetId())
	if err != nil {
		return nil, grpcError(err)
	}

	if deleted {
		return nil, status.Error(codes.FailedPrecondition, "URL is gone")
	}

	// Формируем ответ
	return shortenerpb.URLExpandResponse_builder{Result: orig}.Build(), nil
}

// ListUserURLs — аналог GET /api/user/urls; пустой список допустим (без HTTP 204).
func (s *Server) ListUserURLs(ctx context.Context, _ *shortenerpb.ListUserURLsRequest) (*shortenerpb.UserURLsResponse, error) {
	userID, ok := shortener.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "user id missing in context")
	}

	items, err := shortener.ListURLs(ctx, s.url, userID)
	if err != nil {
		return nil, grpcError(err)
	}

	// Формируем список ссылок
	out := make([]*shortenerpb.URLData, 0, len(items))
	for i := range items {
		out = append(out, shortenerpb.URLData_builder{
			ShortUrl:    items[i].ShortURL,
			OriginalUrl: items[i].OriginalURL,
		}.Build())
	}

	// Формируем ответ
	return shortenerpb.UserURLsResponse_builder{Url: out}.Build(), nil
}
