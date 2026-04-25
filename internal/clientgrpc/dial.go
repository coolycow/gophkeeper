package clientgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// DialOptions управляет транспортной безопасностью для gRPC соединения.
type DialOptions struct {
	Address   string // Address это host:port gRPC сервера.
	Insecure  bool   // Insecure отключает TLS (только для разработки).
	TLSCAFile string // TLSCAFile это опциональный PEM CA bundle; если пустой, используется системный trust store.
}

// Dial открывает клиентское соединение с коротким дефолтным таймаутом для блокирующего handshake.
func Dial(ctx context.Context, opt DialOptions) (*grpc.ClientConn, error) {
	if opt.Address == "" {
		return nil, fmt.Errorf("clientgrpc: empty address")
	}

	// Устанавливаем транспортные credentials
	var tc credentials.TransportCredentials

	// Если Insecure true, используем insecure credentials
	if opt.Insecure {
		tc = insecure.NewCredentials()
	} else {
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if opt.TLSCAFile != "" {
			pemData, err := os.ReadFile(opt.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("read tls ca: %w", err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pemData) {
				return nil, fmt.Errorf("tls ca: no certificates parsed")
			}
			tlsCfg.RootCAs = pool
		}
		tc = credentials.NewTLS(tlsCfg)
	}

	// Устанавливаем таймаут для блокирующего handshake
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return grpc.DialContext(dialCtx, opt.Address,
		grpc.WithTransportCredentials(tc),
	)
}
