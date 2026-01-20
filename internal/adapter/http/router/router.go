package router

import (
	"github.com/blackhorseya/go-ddd/internal/adapter/http/middleware"
	"github.com/blackhorseya/go-ddd/pkg/logx"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Options holds router configuration.
type Options struct {
	Mode   string // gin.DebugMode, gin.ReleaseMode, gin.TestMode
	Logger *logx.Logger
	CORS   cors.Config
}

// DefaultOptions returns default router options.
func DefaultOptions(logger *logx.Logger) Options {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true

	return Options{
		Mode:   gin.ReleaseMode,
		Logger: logger,
		CORS:   corsConfig,
	}
}

// New creates a new Gin router with middleware configured.
func New(opts Options) *gin.Engine {
	gin.SetMode(opts.Mode)

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(cors.New(opts.CORS))

	if opts.Logger != nil {
		r.Use(middleware.Logging(opts.Logger))
	}

	return r
}
