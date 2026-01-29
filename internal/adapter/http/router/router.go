package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/blackhorseya/go-ddd/internal/adapter/http/middleware"

	_ "github.com/blackhorseya/go-ddd/api/openapi" // swagger docs
)

// Options holds router configuration.
type Options struct {
	Mode        string // gin.DebugMode, gin.ReleaseMode, gin.TestMode
	ServiceName string // Service name for tracing
	CORS        cors.Config
}

// DefaultOptions returns default router options.
func DefaultOptions(serviceName string) Options {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true

	return Options{
		Mode:        gin.ReleaseMode,
		ServiceName: serviceName,
		CORS:        corsConfig,
	}
}

// New creates a new Gin router with middleware configured.
func New(opts Options) *gin.Engine {
	gin.SetMode(opts.Mode)

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(cors.New(opts.CORS))
	r.Use(middleware.Tracing(opts.ServiceName))
	r.Use(middleware.TraceID())
	r.Use(middleware.Logging())

	// Swagger documentation
	r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
