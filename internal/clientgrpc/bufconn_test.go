package clientgrpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1024 * 1024

type listAuthStub struct {
	gophkeeperpb.UnimplementedGophKeeperServiceServer
	t *testing.T
}

func (s *listAuthStub) ListSecrets(ctx context.Context, _ *gophkeeperpb.ListSecretsRequest) (*gophkeeperpb.ListSecretsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	require.True(s.t, ok)
	v := md.Get("authorization")
	require.NotEmpty(s.t, v)
	require.Contains(s.t, v[0], "Bearer ")
	return &gophkeeperpb.ListSecretsResponse{}, nil
}

func TestClient_ListSecrets_sendsBearer(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	gophkeeperpb.RegisterGophKeeperServiceServer(srv, &listAuthStub{t: t})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	cli := NewClient(conn)
	require.NoError(t, cli.ApplyAuthResponse(&gophkeeperpb.AuthResponse{
		Token:        "tok",
		RefreshToken: "rt",
		Salt:         "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
		ExpiresAt:    timestamppb.New(time.Now().Add(time.Hour)),
	}))
	_, err = cli.ListSecrets(ctx, gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_ACTIVE_ONLY)
	require.NoError(t, err)
}

type refreshStub struct {
	gophkeeperpb.UnimplementedGophKeeperServiceServer
	listCalls int
}

func (s *refreshStub) ListSecrets(context.Context, *gophkeeperpb.ListSecretsRequest) (*gophkeeperpb.ListSecretsResponse, error) {
	s.listCalls++
	if s.listCalls == 1 {
		return nil, status.Error(codes.Unauthenticated, "expired")
	}
	return &gophkeeperpb.ListSecretsResponse{}, nil
}

func (s *refreshStub) RefreshToken(context.Context, *gophkeeperpb.RefreshTokenRequest) (*gophkeeperpb.AuthResponse, error) {
	return &gophkeeperpb.AuthResponse{
		Token:        "new-access",
		RefreshToken: "new-refresh",
		Salt:         "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
		ExpiresAt:    timestamppb.New(time.Now().Add(time.Hour)),
	}, nil
}

func TestClient_ListSecrets_retriesAfterRefresh(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	st := &refreshStub{}
	gophkeeperpb.RegisterGophKeeperServiceServer(srv, st)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	cli := NewClient(conn)
	require.NoError(t, cli.ApplyAuthResponse(&gophkeeperpb.AuthResponse{
		Token:        "old",
		RefreshToken: "rt-1",
		Salt:         "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
		ExpiresAt:    timestamppb.New(time.Now().Add(time.Hour)),
	}))
	_, err = cli.ListSecrets(ctx, gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_ACTIVE_ONLY)
	require.NoError(t, err)
	require.Equal(t, 2, st.listCalls)
	require.Equal(t, "new-refresh", cli.RefreshToken())
}
