package grpcserver_test

import (
	"context"
	"net"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/grpcserver"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/proto/shortenerpb"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// bufSize — достаточный размер буфера in-memory соединения для одного вызова.
const bufSize = 1024 * 1024

// TestShortenURL_InMemory проверяет цепочку: интерцептор auth → ShortenURL → in-memory repo.
func TestShortenURL_InMemory(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		BaseURL:                           "http://127.0.0.1:8080",
		RandomStringLength:                6,
		RandomStringMaxLength:             32,
		RandomStringMaxGenerationAttempts: 100,
		SecretKey:                         "shortener_secret_key",
	}

	// Пустой путь — без файла на диске, чтобы тест не держал открытый файл под Windows.
	repo := repository.NewDoubleMapsRepository("")
	urlSvc := service.NewURLService(cfg, repo)
	userSvc := service.NewUserService(cfg, repo)

	lis := bufconn.Listen(bufSize)
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthUnaryServerInterceptor(userSvc)))
	grpcserver.NewServer(urlSvc, audit.NewNotifier("", "")).Register(grpcSrv)

	go func() {
		_ = grpcSrv.Serve(lis)
	}()
	t.Cleanup(func() { grpcSrv.Stop() })

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	// Создаём клиента
	client := shortenerpb.NewShortenerServiceClient(conn)

	// Создаём запрос и отправляем его
	resp, err := client.ShortenURL(ctx, shortenerpb.URLShortenRequest_builder{Url: "https://example.com/path"}.Build())
	require.NoError(t, err)
	require.Contains(t, resp.GetResult(), "http://127.0.0.1:8080/")
}
