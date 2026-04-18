// Package router настраивает маршруты и middleware HTTP-сервера.
package router

import (
	"net/http/pprof"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/middleware"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/repository"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт HTTP-роутер с маршрутами сервиса, gzip, логированием и pprof.
// По сути в приложении не используется HTTP-сервер для реальной работы сервиса.
// Но для тестирования и отладки используется HTTP-сервер с маршрутами для pprof.
// Изначально была идея сделать HTTP-сервер для реальной работы сервиса, но в итоге остановился на gRPC.
func NewRouter(cfg *config.ConfigServer, repo repository.GophKeeperRepository, auditNotifier *audit.Notifier) *gin.Engine {
	router := gin.Default()

	// Используем gzip для сжатия данных
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	// Логируем запросы
	router.Use(middleware.RequestLogger())
	// Обрабатываем ошибки
	router.Use(middleware.ErrorHandler())
	// Сжимаем данные
	router.Use(middleware.RequestGzip())

	// Настраиваем маршруты для pprof
	setupPprof(router)

	return router
}

// setupPprof настраивает маршруты для pprof.
func setupPprof(r *gin.Engine) {
	g := r.Group("/debug/pprof")
	g.GET("/", gin.WrapF(pprof.Index))
	g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	g.GET("/profile", gin.WrapF(pprof.Profile))
	g.GET("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/trace", gin.WrapF(pprof.Trace))
	g.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	g.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	g.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	g.GET("/block", gin.WrapH(pprof.Handler("block")))
	g.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	g.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}
