package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coolycow/gophkeeper/internal/buildinfo"
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
	"google.golang.org/grpc/reflection"
)

// Сюда подставляются значения при сборке через -ldflags (см. README.md).
var buildVersion, buildDate, buildCommit string

// init инициализирует информацию о сборке
func init() {
	buildinfo.Version = buildVersion    // ldflags: -X main.buildVersion=...
	buildinfo.BuildDate = buildDate     // ldflags: -X main.buildDate=...
	buildinfo.BuildCommit = buildCommit // ldflags: -X main.buildCommit=...
}

func main() {
	// Вывод информации о сборке при старте
	buildinfo.Fprint(os.Stdout)

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
	auditNotifier, err := audit.NewNotifier(cfg.AuditFile, cfg.AuditURL)

	// Если ошибка при инициализации аудита, выводим ошибку и завершаем программу
	if err != nil {
		log.Fatalf("Failed to initialize audit: %v", err)
	}

	// В конце работы приложения необходимо правильно закрыть аудит (если файл не открыт, то ошибки не будет)
	defer func() {
		if closeErr := auditNotifier.Close(); closeErr != nil {
			logger.Log.Error("Error closing audit notifier", zap.Error(closeErr))
		}
	}()

	// Слой приложения: один раз на процесс (gRPC и при необходимости HTTP-хендлеры используют те же экземпляры).
	userSvc, secretSvc, secretVersionSvc, attachmentSvc := newAppServices(cfg, repo)

	// Инициализируем роутер (сейчас только pprof; общий repo и audit уже передаём для возможного расширения).
	r := router.NewRouter(cfg, repo, auditNotifier)

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

	// Server reflection: grpcurl / Insomnia могут узнавать схему без локального .proto.
	reflection.Register(grpcSrv)

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

// newAppServices создаёт сервисный слой один раз поверх общего repo.
func newAppServices(cfg *config.ConfigServer, repo repository.GophKeeperRepository) (
	service.UserService,
	service.SecretService,
	service.SecretVersionService,
	service.AttachmentService,
) {
	userSvc := service.NewUserService(cfg, repo)
	secretSvc := service.NewSecretService(cfg, repo)
	secretVerSvc := service.NewSecretVersionService(cfg, repo)
	attachmentSvc := service.NewAttachmentService(cfg, repo)
	return userSvc, secretSvc, secretVerSvc, attachmentSvc
}
