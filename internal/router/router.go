// Package router настраивает маршруты и middleware HTTP-сервера.
package router

import (
	"net/http/pprof"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт HTTP-роутер с маршрутами сервиса коротких ссылок, gzip, логированием и pprof.
func NewRouter(cfg *config.Config, repo repository.URLRepository, auditNotifier *audit.Notifier) *gin.Engine {
	router := gin.Default()

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestGzip())

	setupURLRoutes(router, cfg, repo, auditNotifier)

	setupPprof(router)

	return router
}

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
