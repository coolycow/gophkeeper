package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/grpcserver"
	"github.com/coolycow/gophkeeper/internal/logger"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/repository"
	"github.com/coolycow/gophkeeper/internal/router"
	"github.com/coolycow/gophkeeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func main() {
	// Вывод информации о сборке при старте
	fmt.Fprintf(os.Stdout, "Build version: %s\n", orNA(buildVersion))
	fmt.Fprintf(os.Stdout, "Build date: %s\n", orNA(buildDate))
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", orNA(buildCommit))

	// Инициализируем настройки (приоритет: окружение, флаги, дефолт)
	cfg, err := config.InitConfigServer()

	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Инициализируем логер
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Выводим настройки в лог
	cfg.PrintConfig()

	// Инициализируем репозиторий
	var repo repository.GophKeeperRepository
	if cfg.DatabaseDSN != "" {
		repo, err = repository.NewPostgresRepository(cfg.DatabaseDSN)

		if err != nil {
			log.Fatalf("Failed to initialize postgres repository: %v", err)
		}

		logger.Log.Info("Initialized postgres repository", zap.String("dsn", cfg.DatabaseDSN))

		if cfg.RunMigrations {
			if err = repo.RunMigrations(); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			logger.Log.Info("Run migrations succeeded")
			return
		}
	} else {
		log.Fatalf("Database DSN is not set")
	}

	// В конце работы приложения необходимо правильно закрыть хранилище.
	defer func() {
		if err = repo.Close(); err != nil {
			logger.Log.Error("Error closing repository", zap.Error(err))
		}
	}()

	// Инициализируем аудит
	auditNotifier := audit.NewNotifier(cfg.AuditFile, cfg.AuditURL)

	// Инициализируем роутер
	r := router.NewRouter(cfg, repo, auditNotifier)

	// Сервисы для gRPC совпадают по смыслу с теми, что создаёт router (общий repo).
	userSvc := service.NewUserService(cfg, repo)
	secretSvc := service.NewSecretService(cfg, repo)
	secretVersionSvc := service.NewSecretVersionService(cfg, repo)
	attachmentSvc := service.NewAttachmentService(cfg, repo)

	// Получаем адрес сервера из настроек и запускаем сервер
	serverAddress := cfg.GetServerAddress()
	logger.Log.Info("Running server ", zap.String("address", serverAddress))

	// Инициализируем сервер
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	// Запускаем сервер
	go func() {
		var serveErr error
		if cfg.EnableHTTPS {
			serveErr = srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			serveErr = srv.ListenAndServe()
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Log.Fatal("server error", zap.Error(serveErr))
		}
	}()

	// gRPC: тот же Host, отдельный порт (GrpcPort или Port+1), TLS при EnableHTTPS.
	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcserver.AuthUnaryServerInterceptor(userSvc)),
	}
	if cfg.EnableHTTPS {
		creds, tlsErr := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if tlsErr != nil {
			log.Fatalf("Failed to load gRPC TLS credentials: %v", tlsErr)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// Инициализируем gRPC сервер
	grpcSrv := grpc.NewServer(grpcOpts...)
	grpcserver.NewServer(auditNotifier, userSvc, secretSvc, secretVersionSvc, attachmentSvc).RegisterGRPC(grpcSrv)

	// Получаем адрес gRPC сервера и запускаем сервер
	grpcLis, err := net.Listen("tcp", cfg.GetGRPCServerAddress())
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	logger.Log.Info("Running gRPC server", zap.String("address", cfg.GetGRPCServerAddress()))

	// Запускаем gRPC сервер
	go func() {
		if serveErr := grpcSrv.Serve(grpcLis); serveErr != nil {
			logger.Log.Fatal("gRPC server error", zap.Error(serveErr))
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit
	logger.Log.Info("shutdown signal received", zap.String("signal", sig.String()))

	// Создаем контекст для завершения работы сервера
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Параллельно завершаем HTTP и gRPC.
	var shutdownWg sync.WaitGroup
	shutdownWg.Add(2)

	go func() {
		defer shutdownWg.Done()
		grpcSrv.GracefulStop()
	}()

	go func() {
		defer shutdownWg.Done()
		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Log.Error("graceful shutdown failed", zap.Error(shutdownErr))
		}
	}()

	shutdownWg.Wait()
}
