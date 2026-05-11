package grpcserver

import (
	"context"
	"net"
	"testing"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/coolycow/gophkeeper/internal/repository/repotest"
	"github.com/coolycow/gophkeeper/internal/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSz = 1 << 20

// Контрольная проверка: Ping проходит через зарегистрированный Server (без БД — заглушка repo).
func TestSmoke_Ping(t *testing.T) {
	repo := &repotest.Stub{}
	cfg := &config.ConfigServer{}
	auditN, err := audit.NewNotifier("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = auditN.Close() })

	srv := NewServer(auditN,
		service.NewUserService(cfg, repo),
		service.NewSecretService(cfg, repo),
		service.NewSecretVersionService(cfg, repo),
		service.NewAttachmentService(cfg, repo),
	)

	grpcSrv := grpc.NewServer()
	srv.RegisterGRPC(grpcSrv)
	lis := bufconn.Listen(bufSz)
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(func() { grpcSrv.Stop() })

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "buf",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	cli := gophkeeperpb.NewGophKeeperServiceClient(conn)
	resp, err := cli.Ping(ctx, &gophkeeperpb.PingRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
